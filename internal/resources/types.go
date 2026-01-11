package resources

type ResourceDefinition struct {
	CPU     CPUCoreInput
	Memory  MemoryInput
	Storage StorageInput
	GPU     *GPUDefinition
}

type GPUDefinition struct {
	Vendor string // nvidia, intel, amd, or any.
	Count  GPUCountInput
	Memory GPUMemoryInput
}

type CPUCoreInput int
type CPUCore uint64

func (m *CPUCoreInput) Normalize() CPUCore {
	return CPUCore(*m)
}

type MemoryInput string
type MemoryMB uint64

func (m *MemoryInput) Normalize() MemoryMB {
	tmp, err := parseMemMB(string(*m))
	if err != nil {
		return 0
	}

	return tmp
}

type StorageInput string
type StorageMB uint64

func (m *StorageInput) Normalize() StorageMB {
	tmp, err := parseStorageMB(string(*m))
	if err != nil {
		return 0
	}

	return tmp
}

type GPUCountInput int
type GPUCount int

func (m *GPUCountInput) Normalize() GPUCount {
	return GPUCount(*m)
}

type GPUMemoryInput string
type GPUMemoryMB int

func (m *GPUMemoryInput) Normalize() GPUMemoryMB {
	tmp, err := parseUnitStringToMB(string(*m))
	if err != nil {
		return 0
	}

	return GPUMemoryMB(tmp)
}
