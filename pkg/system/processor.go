package system

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type Processor int

const (
	ProcessorUnknown Processor = iota
	ProcessorX86
	ProcessorX64
	ProcessorArm32
	ProcessorArm64
	ProcessorRISCV32
	ProcessorRISCV64
)

func (p Processor) String() string {
	switch p {
	case ProcessorX86:
		return "x86"
	case ProcessorX64:
		return "x64"
	case ProcessorArm32:
		return "arm"
	case ProcessorArm64:
		return "arm64"
	case ProcessorRISCV32:
		return "riscv32"
	case ProcessorRISCV64:
		return "riscv64"
	}
	return "unknown"
}

func (p Processor) StringLower() string {
	return strings.ToLower(p.String())
}

func (p Processor) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

func (p *Processor) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "x86", "i386", "i686":
		*p = ProcessorX86
	case "x64", "x86_64", "amd64":
		*p = ProcessorX64
	case "arm", "arm32", "armv7l":
		*p = ProcessorArm32
	case "arm64", "aarch64":
		*p = ProcessorArm64
	case "riscv32":
		*p = ProcessorRISCV32
	case "riscv64":
		*p = ProcessorRISCV64
	default:
		*p = ProcessorUnknown
	}
	return nil
}
