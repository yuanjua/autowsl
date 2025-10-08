package system

import (
	"runtime"
	"strings"
)

// HostArchitecture represents the system architecture
type HostArchitecture string

const (
	ArchX64   HostArchitecture = "x64"
	ArchARM64 HostArchitecture = "arm64"
	ArchX86   HostArchitecture = "x86"
)

// GetHostArchitecture returns the current system architecture
func GetHostArchitecture() HostArchitecture {
	arch := runtime.GOARCH

	switch arch {
	case "amd64":
		return ArchX64
	case "arm64":
		return ArchARM64
	case "386":
		return ArchX86
	default:
		// Default to x64 for unknown architectures
		return ArchX64
	}
}

// IsCompatibleArchitecture checks if a distribution architecture is compatible with host
func IsCompatibleArchitecture(distroArch string) bool {
	hostArch := GetHostArchitecture()
	distroArchLower := strings.ToLower(distroArch)

	switch hostArch {
	case ArchX64:
		// x64 can run x64 and x86
		return strings.Contains(distroArchLower, "x64") ||
			strings.Contains(distroArchLower, "x86") ||
			strings.Contains(distroArchLower, "amd64")
	case ArchARM64:
		// ARM64 can run ARM64 (and potentially ARM32)
		return strings.Contains(distroArchLower, "arm64") ||
			strings.Contains(distroArchLower, "aarch64")
	case ArchX86:
		// x86 can only run x86
		return strings.Contains(distroArchLower, "x86") &&
			!strings.Contains(distroArchLower, "x64")
	default:
		return false
	}
}

// GetPreferredArchitectureSuffix returns the preferred architecture suffix for filtering
func GetPreferredArchitectureSuffix() string {
	hostArch := GetHostArchitecture()

	switch hostArch {
	case ArchX64:
		return "x64"
	case ArchARM64:
		return "arm64"
	case ArchX86:
		return "x86"
	default:
		return "x64"
	}
}

// ShouldSkipArchitecture returns true if the architecture should be skipped
func ShouldSkipArchitecture(filename string) bool {
	hostArch := GetHostArchitecture()
	filenameLower := strings.ToLower(filename)

	switch hostArch {
	case ArchX64:
		// Skip ARM versions on x64 systems
		return strings.Contains(filenameLower, "arm64") ||
			strings.Contains(filenameLower, "arm32") ||
			strings.Contains(filenameLower, "arm_") ||
			strings.Contains(filenameLower, "aarch64")
	case ArchARM64:
		// Skip x64 versions on ARM64 systems
		return strings.Contains(filenameLower, "x64") ||
			strings.Contains(filenameLower, "amd64") ||
			strings.Contains(filenameLower, "_64") && !strings.Contains(filenameLower, "arm64")
	default:
		return false
	}
}
