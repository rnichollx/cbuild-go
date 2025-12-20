package host

import (
	"runtime"
)

// DetectHostPlatform returns the host platform name as used in toolchain definitions
func DetectHostPlatform() string {
	os := runtime.GOOS
	switch os {
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	default:
		return os
	}
}

// DetectHostArch returns the host architecture name as used in toolchain definitions
func DetectHostArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		{
			return "x64"
		}
	case "386":
		{
			return "x86"
		}
	case "arm64":
		{
			return "arm64"
		}
	case "arm":
		{
			return "arm32"
		}
	default:
		return arch
	}
}
