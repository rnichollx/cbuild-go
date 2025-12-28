package cmake

import (
	"gitlab.com/rpnx/cbuild-go/pkg/system"
	"testing"
)

func TestProcessorToCMakeName(t *testing.T) {
	tests := []struct {
		platform  system.Platform
		processor system.Processor
		want      string
		wantErr   bool
	}{
		{system.PlatformLinux, system.ProcessorX86, "i686", false},
		{system.PlatformLinux, system.ProcessorX64, "x86_64", false},
		{system.PlatformLinux, system.ProcessorArm64, "aarch64", false},
		{system.PlatformMac, system.ProcessorArm64, "arm64", false},
		{system.PlatformMac, system.ProcessorX64, "x86_64", false},
		{system.PlatformWindows, system.ProcessorX64, "AMD64", false},
		{system.PlatformWindows, system.ProcessorArm64, "ARM64", false},
		{system.PlatformFreeBSD, system.ProcessorX64, "amd64", false},
		{system.PlatformFreeBSD, system.ProcessorX86, "i386", false},
		{system.PlatformLinux, system.ProcessorUnknown, "", true},
	}

	for _, tt := range tests {
		got, err := ProcessorToCMakeName(tt.platform, tt.processor)
		if (err != nil) != tt.wantErr {
			t.Errorf("ProcessorToCMakeName(%v, %v) error = %v, wantErr %v", tt.platform, tt.processor, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ProcessorToCMakeName(%v, %v) = %v, want %v", tt.platform, tt.processor, got, tt.want)
		}
	}
}
