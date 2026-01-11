// internal/models/resources/standard_tiers.go

package resources

var (
	// ==================== CPU-Optimized (C-series) ====================
	C_Nano = ResourceDefinition{
		CPU:     1,
		Memory:  "512mb",
		Storage: "10gb",
	}

	C_Micro = ResourceDefinition{
		CPU:     1,
		Memory:  "1gb",
		Storage: "10gb",
	}

	C_Small = ResourceDefinition{
		CPU:     2,
		Memory:  "2gb",
		Storage: "20gb",
	}

	C_Medium = ResourceDefinition{
		CPU:     4,
		Memory:  "4gb",
		Storage: "50gb",
	}

	C_Large = ResourceDefinition{
		CPU:     8,
		Memory:  "8gb",
		Storage: "100gb",
	}

	C_XLarge = ResourceDefinition{
		CPU:     16,
		Memory:  "16gb",
		Storage: "250gb",
	}

	C_2XLarge = ResourceDefinition{
		CPU:     32,
		Memory:  "32gb",
		Storage: "500gb",
	}

	C_4XLarge = ResourceDefinition{
		CPU:     64,
		Memory:  "64gb",
		Storage: "1tb",
	}

	// ==================== Memory-Optimized (M-series) ====================
	M_Small = ResourceDefinition{
		CPU:     2,
		Memory:  "4gb",
		Storage: "20gb",
	}

	M_Medium = ResourceDefinition{
		CPU:     4,
		Memory:  "16gb",
		Storage: "50gb",
	}

	M_Large = ResourceDefinition{
		CPU:     8,
		Memory:  "32gb",
		Storage: "100gb",
	}

	M_XLarge = ResourceDefinition{
		CPU:     16,
		Memory:  "64gb",
		Storage: "250gb",
	}

	M_2XLarge = ResourceDefinition{
		CPU:     32,
		Memory:  "128gb",
		Storage: "500gb",
	}

	M_4XLarge = ResourceDefinition{
		CPU:     64,
		Memory:  "256gb",
		Storage: "1tb",
	}

	// ==================== GPU Generic (G-series - any vendor) ====================
	G_Micro = ResourceDefinition{
		CPU:     4,
		Memory:  "8gb",
		Storage: "50gb",
		GPU: &GPUDefinition{
			Vendor: "any",
			Count:  1,
			Memory: "2gb",
		},
	}

	G_Small = ResourceDefinition{ // RTX 3060, T4, RX 6600
		CPU:     4,
		Memory:  "16gb",
		Storage: "100gb",
		GPU: &GPUDefinition{
			Vendor: "any",
			Count:  1,
			Memory: "8gb",
		},
	}

	G_Medium = ResourceDefinition{ // RTX 4080, RTX A4000, RX 7900 XT
		CPU:     8,
		Memory:  "32gb",
		Storage: "250gb",
		GPU: &GPUDefinition{
			Vendor: "any",
			Count:  1,
			Memory: "16gb",
		},
	}

	G_Large = ResourceDefinition{ // RTX 4090, A5000, L40, RX 7900 XTX
		CPU:     12,
		Memory:  "48gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "any",
			Count:  1,
			Memory: "24gb",
		},
	}

	G_XLarge = ResourceDefinition{ // A100 40GB, A6000, MI100
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "any",
			Count:  1,
			Memory: "40gb",
		},
	}

	G_2XLarge = ResourceDefinition{ // A100 80GB, H100, MI250X
		CPU:     24,
		Memory:  "96gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "any",
			Count:  1,
			Memory: "80gb",
		},
	}

	// ==================== GPU NVIDIA (GN-series) ====================
	GN_Small = ResourceDefinition{ // T4, RTX 3060, RTX 3060 Ti
		CPU:     4,
		Memory:  "16gb",
		Storage: "100gb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  1,
			Memory: "8gb",
		},
	}

	GN_Medium = ResourceDefinition{ // RTX 4080, RTX A4000, RTX A5000
		CPU:     8,
		Memory:  "32gb",
		Storage: "250gb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  1,
			Memory: "16gb",
		},
	}

	GN_Large = ResourceDefinition{ // RTX 4090, L40, L40S, RTX 6000 Ada
		CPU:     12,
		Memory:  "48gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  1,
			Memory: "24gb",
		},
	}

	GN_XLarge = ResourceDefinition{ // A100 40GB, A6000
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  1,
			Memory: "40gb",
		},
	}

	GN_2XLarge = ResourceDefinition{ // A100 80GB, H100, H200
		CPU:     24,
		Memory:  "96gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  1,
			Memory: "80gb",
		},
	}

	// ==================== GPU AMD (GA-series) ====================
	GA_Small = ResourceDefinition{ // RX 6600 XT, RX 6700 XT
		CPU:     4,
		Memory:  "16gb",
		Storage: "100gb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  1,
			Memory: "8gb",
		},
	}

	GA_Medium = ResourceDefinition{ // RX 7900 XT, W6800
		CPU:     8,
		Memory:  "32gb",
		Storage: "250gb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  1,
			Memory: "16gb",
		},
	}

	GA_Large = ResourceDefinition{ // RX 7900 XTX, W7900
		CPU:     12,
		Memory:  "48gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  1,
			Memory: "24gb",
		},
	}

	GA_XLarge = ResourceDefinition{ // MI100, Radeon Pro W7900
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  1,
			Memory: "32gb",
		},
	}

	GA_2XLarge = ResourceDefinition{ // MI210, MI250 (per GCD), MI250X
		CPU:     24,
		Memory:  "96gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  1,
			Memory: "64gb",
		},
	}

	// ==================== GPU Apple Silicon (GM-series - Metal) ====================
	GM_Small = ResourceDefinition{ // M1, M2 (base)
		CPU:     8,
		Memory:  "16gb", // Unified memory
		Storage: "250gb",
		GPU: &GPUDefinition{
			Vendor: "apple",
			Count:  1,
			Memory: "16gb", // Shared with system RAM
		},
	}

	GM_Medium = ResourceDefinition{ // M1 Pro, M2 Pro
		CPU:     10,
		Memory:  "32gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "apple",
			Count:  1,
			Memory: "32gb",
		},
	}

	GM_Large = ResourceDefinition{ // M1 Max, M2 Max
		CPU:     12,
		Memory:  "48gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "apple",
			Count:  1,
			Memory: "48gb",
		},
	}

	GM_XLarge = ResourceDefinition{ // M1 Ultra, M2 Ultra
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "apple",
			Count:  1,
			Memory: "64gb",
		},
	}

	GM_2XLarge = ResourceDefinition{ // M2 Ultra (maxed), M3 Ultra (future)
		CPU:     24,
		Memory:  "128gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "apple",
			Count:  1,
			Memory: "128gb",
		},
	}

	// ==================== GPU Intel (GI-series) ====================
	GI_Small = ResourceDefinition{ // Arc A380, Arc A580
		CPU:     4,
		Memory:  "16gb",
		Storage: "100gb",
		GPU: &GPUDefinition{
			Vendor: "intel",
			Count:  1,
			Memory: "8gb",
		},
	}

	GI_Medium = ResourceDefinition{ // Arc A770, Arc A750
		CPU:     8,
		Memory:  "32gb",
		Storage: "250gb",
		GPU: &GPUDefinition{
			Vendor: "intel",
			Count:  1,
			Memory: "16gb",
		},
	}

	GI_Large = ResourceDefinition{ // Data Center GPU Max 1100
		CPU:     12,
		Memory:  "48gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "intel",
			Count:  1,
			Memory: "48gb",
		},
	}

	GI_XLarge = ResourceDefinition{ // Data Center GPU Max 1550
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "intel",
			Count:  1,
			Memory: "64gb",
		},
	}

	// ==================== Multi-GPU NVIDIA (PN-series) ====================
	PN_2XLarge = ResourceDefinition{ // 2x RTX 4090, 2x L40
		CPU:     16,
		Memory:  "64gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  2,
			Memory: "24gb",
		},
	}

	PN_4XLarge = ResourceDefinition{ // 4x L40, 4x RTX 4090
		CPU:     32,
		Memory:  "128gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  4,
			Memory: "24gb",
		},
	}

	PN_8XLarge = ResourceDefinition{ // DGX A100 (40GB variant)
		CPU:     64,
		Memory:  "256gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  8,
			Memory: "24gb",
		},
	}

	// ==================== Multi-GPU AMD (PA-series) ====================
	PA_2XLarge = ResourceDefinition{ // 2x RX 7900 XTX, 2x W7900
		CPU:     16,
		Memory:  "64gb",
		Storage: "500gb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  2,
			Memory: "24gb",
		},
	}

	PA_4XLarge = ResourceDefinition{ // 4x MI210, 4x W7900
		CPU:     32,
		Memory:  "128gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  4,
			Memory: "24gb",
		},
	}

	PA_8XLarge = ResourceDefinition{ // 8x MI210, 4x MI250X (dual-GCD)
		CPU:     64,
		Memory:  "256gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  8,
			Memory: "32gb",
		},
	}

	// ==================== High-Memory NVIDIA (HN-series) ====================
	HN_XLarge = ResourceDefinition{ // 1x A100 80GB, 1x H100
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  1,
			Memory: "80gb",
		},
	}

	HN_2XLarge = ResourceDefinition{ // 2x A100 80GB, 2x H100
		CPU:     32,
		Memory:  "128gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  2,
			Memory: "80gb",
		},
	}

	HN_4XLarge = ResourceDefinition{ // 4x A100 80GB, 4x H100
		CPU:     64,
		Memory:  "256gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  4,
			Memory: "80gb",
		},
	}

	HN_8XLarge = ResourceDefinition{ // DGX A100 80GB, DGX H100
		CPU:     128,
		Memory:  "512gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "nvidia",
			Count:  8,
			Memory: "80gb",
		},
	}

	// ==================== High-Memory AMD (HA-series) ====================
	HA_XLarge = ResourceDefinition{ // 1x MI250X
		CPU:     16,
		Memory:  "64gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  1,
			Memory: "64gb",
		},
	}

	HA_2XLarge = ResourceDefinition{ // 2x MI250X
		CPU:     32,
		Memory:  "128gb",
		Storage: "1tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  2,
			Memory: "64gb",
		},
	}

	HA_4XLarge = ResourceDefinition{ // 4x MI250X
		CPU:     64,
		Memory:  "256gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  4,
			Memory: "64gb",
		},
	}

	HA_8XLarge = ResourceDefinition{ // 8x MI250X
		CPU:     128,
		Memory:  "512gb",
		Storage: "2tb",
		GPU: &GPUDefinition{
			Vendor: "amd",
			Count:  8,
			Memory: "64gb",
		},
	}
)

// StandardTierMap maps tier names to their definitions
var StandardTierMap = map[string]ResourceDefinition{
	// CPU-Optimized
	"c-nano":    C_Nano,
	"c-micro":   C_Micro,
	"c-small":   C_Small,
	"c-medium":  C_Medium,
	"c-large":   C_Large,
	"c-xlarge":  C_XLarge,
	"c-2xlarge": C_2XLarge,
	"c-4xlarge": C_4XLarge,

	// Memory-Optimized
	"m-small":   M_Small,
	"m-medium":  M_Medium,
	"m-large":   M_Large,
	"m-xlarge":  M_XLarge,
	"m-2xlarge": M_2XLarge,
	"m-4xlarge": M_4XLarge,

	// GPU Generic
	"g-micro":   G_Micro,
	"g-small":   G_Small,
	"g-medium":  G_Medium,
	"g-large":   G_Large,
	"g-xlarge":  G_XLarge,
	"g-2xlarge": G_2XLarge,

	// GPU NVIDIA
	"gn-small":   GN_Small,
	"gn-medium":  GN_Medium,
	"gn-large":   GN_Large,
	"gn-xlarge":  GN_XLarge,
	"gn-2xlarge": GN_2XLarge,

	// GPU AMD
	"ga-small":   GA_Small,
	"ga-medium":  GA_Medium,
	"ga-large":   GA_Large,
	"ga-xlarge":  GA_XLarge,
	"ga-2xlarge": GA_2XLarge,

	// GPU Apple Silicon
	"gm-small":   GM_Small,
	"gm-medium":  GM_Medium,
	"gm-large":   GM_Large,
	"gm-xlarge":  GM_XLarge,
	"gm-2xlarge": GM_2XLarge,

	// GPU Intel
	"gi-small":  GI_Small,
	"gi-medium": GI_Medium,
	"gi-large":  GI_Large,
	"gi-xlarge": GI_XLarge,

	// // Multi-GPU NVIDIA
	// "pn-2xlarge": PN_2XLarge,
	// "pn-4xlarge": PN_4XLarge,
	// "pn-8xlarge": PN_8XLarge,

	// // Multi-GPU AMD
	// "pa-2xlarge": PA_2XLarge,
	// "pa-4xlarge": PA_4XLarge,
	// "pa-8xlarge": PA_8XLarge,

	// // High-Memory NVIDIA
	// "hn-xlarge":  HN_XLarge,
	// "hn-2xlarge": HN_2XLarge,
	// "hn-4xlarge": HN_4XLarge,
	// "hn-8xlarge": HN_8XLarge,

	// // High-Memory AMD
	// "ha-xlarge":  HA_XLarge,
	// "ha-2xlarge": HA_2XLarge,
	// "ha-4xlarge": HA_4XLarge,
	// "ha-8xlarge": HA_8XLarge,
}
