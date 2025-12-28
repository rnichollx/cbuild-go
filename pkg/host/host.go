package host

import (
	"runtime"

	"gitlab.com/rpnx/cbuild-go/pkg/system"
)

// DetectHostPlatform returns the host platform name as used in toolchain definitions
func DetectHostPlatform() system.Platform {
	os := runtime.GOOS
	switch os {
	case "linux":
		return system.PlatformLinux
	case "windows":
		return system.PlatformUnknown
	case "darwin":
		return system.PlatformMac
	default:
		return system.PlatformUnknown
	}
}

// DetectHostArch returns the host architecture name as used in toolchain definitions
func DetectHostProcessor() system.Processor {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		{
			return system.ProcessorX64
		}
	case "386":
		{
			return system.ProcessorX86
		}
	case "arm64":
		{
			return system.ProcessorArm64
		}
	case "arm":
		{
			return system.ProcessorArm32
		}
	default:
		return system.ProcessorUnknown
	}
}
