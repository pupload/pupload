package resources

import (
	"slices"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/jaypipes/ghw"
	"github.com/moby/moby/api/types/container"
)

type GPUInfo struct {
	vendor string
	memory uint64
}

func getDeviceRequest(g GPUInfo) []container.DeviceRequest {
	var dr []container.DeviceRequest
	if g.vendor == "nvidia" {
		dr = []container.DeviceRequest{{
			Driver:       "nvidia",
			Count:        1,
			Capabilities: [][]string{{"all"}},
		}}
	}
	if g.vendor == "amd" {
		dr = []container.DeviceRequest{{
			Driver:       "amd",
			Count:        1,
			Capabilities: [][]string{{"gpu"}},
		}}
	}

	return dr
}

func detectGPUResources() ([]GPUInfo, error) {
	info, err := ghw.GPU()
	if err != nil {
		return nil, err
	}

	gpus := make([]GPUInfo, 0)

	for _, g := range info.GraphicsCards {
		gInfo := classifyGPU(g)
		if gInfo == nil {
			continue
		}

		gpus = append(gpus, *gInfo)
	}

	return gpus, nil
}

func classifyGPU(card *ghw.GraphicsCard) *GPUInfo {
	driver := card.DeviceInfo.Driver
	vendor := strings.ToLower(card.DeviceInfo.Vendor.Name)

	if driver == "nvidia" || strings.Contains(strings.ToLower(vendor), "nvidia") {
		return &GPUInfo{
			vendor: "nvidia",
			memory: getNVIDIAMemory(card.DeviceInfo.Address),
		}
	}

	return nil
}

func getNVIDIAMemory(pciAddr string) uint64 {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {

	}
	defer nvml.Shutdown()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {

	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		pci, ret := device.GetPciInfo()
		if ret != nvml.SUCCESS {
			continue
		}

		nullTerminator := slices.Index(pci.BusIdLegacy[:], 0)
		if nullTerminator == -1 {
			nullTerminator = len(pci.BusId)
		}

		addr := string(pci.BusIdLegacy[:nullTerminator])

		if strings.EqualFold(addr, pciAddr) {
			memory, ret := device.GetMemoryInfo()
			if ret != nvml.SUCCESS {
				continue
			}

			return memory.Total / (1024 * 1024)
		}
	}

	return 0
}
