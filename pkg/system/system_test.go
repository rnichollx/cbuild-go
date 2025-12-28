package system

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPlatformYAML(t *testing.T) {
	tests := []struct {
		yaml string
		want Platform
	}{
		{"windows", PlatformWindows},
		{"Windows", PlatformWindows},
		{"mac", PlatformMac},
		{"macos", PlatformMac},
		{"darwin", PlatformMac},
		{"linux", PlatformLinux},
		{"freebsd", PlatformFreeBSD},
		{"unknown", PlatformUnknown},
	}

	for _, tt := range tests {
		var p Platform
		err := yaml.Unmarshal([]byte(tt.yaml), &p)
		if err != nil {
			t.Errorf("yaml.Unmarshal(%q) error = %v", tt.yaml, err)
			continue
		}
		if p != tt.want {
			t.Errorf("yaml.Unmarshal(%q) = %v, want %v", tt.yaml, p, tt.want)
		}

		// Test Marshalling
		data, err := yaml.Marshal(p)
		if err != nil {
			t.Errorf("yaml.Marshal(%v) error = %v", p, err)
			continue
		}
		// Note: Marshalling uses .String() which might be capitalized
		var p2 Platform
		err = yaml.Unmarshal(data, &p2)
		if err != nil {
			t.Errorf("yaml.Unmarshal(yaml.Marshal(%v)) error = %v", p, err)
			continue
		}
		if p2 != p {
			t.Errorf("yaml.Unmarshal(yaml.Marshal(%v)) = %v, want %v", p, p2, p)
		}
	}
}

func TestProcessorYAML(t *testing.T) {
	tests := []struct {
		yaml string
		want Processor
	}{
		{"x86", ProcessorX86},
		{"i386", ProcessorX86},
		{"i686", ProcessorX86},
		{"x64", ProcessorX64},
		{"x86_64", ProcessorX64},
		{"amd64", ProcessorX64},
		{"arm", ProcessorArm32},
		{"arm32", ProcessorArm32},
		{"armv7l", ProcessorArm32},
		{"arm64", ProcessorArm64},
		{"aarch64", ProcessorArm64},
		{"riscv32", ProcessorRISCV32},
		{"riscv64", ProcessorRISCV64},
		{"unknown", ProcessorUnknown},
	}

	for _, tt := range tests {
		var p Processor
		err := yaml.Unmarshal([]byte(tt.yaml), &p)
		if err != nil {
			t.Errorf("yaml.Unmarshal(%q) error = %v", tt.yaml, err)
			continue
		}
		if p != tt.want {
			t.Errorf("yaml.Unmarshal(%q) = %v, want %v", tt.yaml, p, tt.want)
		}

		// Test Marshalling
		data, err := yaml.Marshal(p)
		if err != nil {
			t.Errorf("yaml.Marshal(%v) error = %v", p, err)
			continue
		}
		var p2 Processor
		err = yaml.Unmarshal(data, &p2)
		if err != nil {
			t.Errorf("yaml.Unmarshal(yaml.Marshal(%v)) error = %v", p, err)
			continue
		}
		if p2 != p {
			t.Errorf("yaml.Unmarshal(yaml.Marshal(%v)) = %v, want %v", p, p2, p)
		}
	}
}
