package resources

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"

	"github.com/jaypipes/ghw"
	"github.com/moby/moby/api/types/container"
	"github.com/pupload/pupload/internal/logging"
	"github.com/shirou/gopsutil/v3/disk"
)

type ResourceManager struct {
	MaxCPU       CPUCore
	MaxMemoryMB  MemoryMB
	MaxStorageMB StorageMB

	currMemMB     MemoryMB
	currStorageMB StorageMB

	gpus []GPUInfo

	log *slog.Logger
}

type GPUResources struct {
	Vendor      string
	Features    []string
	MaxCount    int
	MaxMemoryGB float64
}

type ResourceSettings struct {
	MaxCPU     string // # of cores: 1, 2 etc. or auto
	MaxMemory  string // 1G, 512MB, etc. or auto
	MaxStorage string // 1G, 512MB, etc. or auto

	AllowedGPUs []string // pcie of allowed GPU's
}

func CreateResourceManager(cfg ResourceSettings) (*ResourceManager, error) {

	var cpu CPUCore
	var mem MemoryMB
	var sto StorageMB

	if cfg.MaxCPU == "auto" {
		c, err := ghw.CPU()
		if err != nil {
			return nil, err
		}
		cpu = CPUCore(c.TotalHardwareThreads)
	} else {
		tmp, err := parseCPUCore(cfg.MaxCPU)
		if err != nil {
			return nil, err
		}

		cpu = tmp
	}

	if cfg.MaxMemory == "auto" {
		memory, err := ghw.Memory()
		if err != nil {
			return nil, err
		}

		mem = MemoryMB(memory.TotalPhysicalBytes / (1024 * 1024))
	} else {
		tmp, err := parseMemMB(cfg.MaxMemory)
		if err != nil {
			return nil, fmt.Errorf("unable to parse max memory: %w", err)
		}
		mem = tmp
	}

	if cfg.MaxStorage == "auto" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		storage, err := disk.Usage(dir)
		if err != nil {
			return nil, err
		}

		sto = StorageMB(storage.Free / (1024 * 1024))
	} else {
		tmp, err := parseStorageMB(cfg.MaxStorage)
		if err != nil {
			return nil, fmt.Errorf("unable to parse max storage: %w", err)
		}
		sto = tmp
	}

	gpus, err := detectGPUResources()
	if err != nil {
		gpus = []GPUInfo{}
	}

	return &ResourceManager{
		MaxCPU:       cpu,
		MaxMemoryMB:  mem,
		MaxStorageMB: sto,

		gpus: gpus,

		log: logging.ForService("resource-manager"),
	}, nil
}

func parseStorageMB(s string) (StorageMB, error) {
	val, err := parseUnitStringToMB(s)
	return StorageMB(val), err
}

func parseMemMB(s string) (MemoryMB, error) {
	val, err := parseUnitStringToMB(s)
	return MemoryMB(val), err
}

var suffix map[string]float64 = map[string]float64{
	"b":  1.0 / 1024 / 1024,
	"kb": 1.0 / 1024,
	"mb": 1,
	"gb": 1024,
	"tb": 1024 * 1024,
}

func parseUnitStringToMB(s string) (uint64, error) {
	exp, err := regexp.Compile("([0-9]+)(tb|gb|mb|kb|b)")
	if err != nil {
		return 0, err
	}

	groups := exp.FindStringSubmatch(s)
	if len(groups) != 3 {
		return 0, fmt.Errorf("invalid string format: %s", s)
	}

	val, err := strconv.ParseUint(groups[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse value %d: %w", val, err)
	}

	unit := groups[2]
	factor, ok := suffix[unit]
	if !ok {
		return 0, fmt.Errorf("invalid unit suffix %s", unit)
	}

	return uint64(float64(val) * factor), nil

}

func parseCPUCore(s string) (CPUCore, error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return CPUCore(val), err
}

func (rm *ResourceManager) Reserve(tierName string) error {
	resource, ok := StandardTierMap[tierName]
	if !ok {
		return fmt.Errorf("tier not found")
	}

	rm.currMemMB += resource.Memory.Normalize()
	rm.currStorageMB += resource.Storage.Normalize()

	return nil
}

func (rm *ResourceManager) Release(tierName string) error {
	resource, ok := StandardTierMap[tierName]
	if !ok {
		return fmt.Errorf("tier not found")
	}

	rm.currMemMB -= resource.Memory.Normalize()
	rm.currStorageMB -= resource.Storage.Normalize()

	return nil
}

func (rm *ResourceManager) GetValidTierMap() map[string]int {

	validTiers := make(map[string]int)

	validMem := rm.MaxMemoryMB - MemoryMB(rm.currMemMB)
	validStorage := rm.MaxStorageMB - StorageMB(rm.currStorageMB)

	i := 2
	for name, resource := range StandardTierMap {
		cpuValid := rm.MaxCPU >= resource.CPU.Normalize()
		memValid := validMem >= resource.Memory.Normalize()
		storageValid := validStorage >= resource.Storage.Normalize()
		gpuValid := resource.GPU == nil
		for _, gpu := range rm.gpus {
			gpuValid = gpuValid || ((resource.GPU.Vendor == gpu.vendor || resource.GPU.Vendor == "any") && resource.GPU.Memory.Normalize() <= GPUMemoryMB(gpu.memory))
		}

		if cpuValid && memValid && storageValid && gpuValid {
			validTiers[name] = i
			i++
		}
	}

	validTiers["worker"] = 1

	rm.log.Debug("currently allowed tiers", "tiers", validTiers)

	return validTiers
}

func (rm *ResourceManager) GenerateContainerResource(tierName string) (container.Resources, error) {
	r, ok := StandardTierMap[tierName]
	if !ok {
		return container.Resources{}, fmt.Errorf("invalid tier name %s", tierName)
	}

	var dr []container.DeviceRequest = nil
	if r.GPU != nil && len(rm.gpus) != 0 {
		dr = getDeviceRequest(rm.gpus[0])
	}

	resource := container.Resources{
		NanoCPUs:          int64(r.CPU.Normalize()),
		MemoryReservation: int64(r.Memory.Normalize()) * 1024 * 1024,
		DeviceRequests:    dr,
	}

	return resource, nil
}
