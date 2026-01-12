package cmake

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/system"
	"gopkg.in/yaml.v3"
)

type GenerateToolchainFileOptions struct {
	CCompiler       string
	CXXCompiler     string
	Linker          string
	ExtraCXXFlags   []string
	SystemPlatform  system.Platform
	SystemProcessor system.Processor
	WorkspaceDir    string
	OutputFile      string
}

func PlatformToCMakeName(platform system.Platform) (string, error) {
	switch platform {
	case system.PlatformMac:
		return "Darwin", nil
	case system.PlatformLinux:
		return "Linux", nil
	case system.PlatformFreeBSD:
		return "FreeBSD", nil
	case system.PlatformWindows:
		return "Windows", nil
	default:
		return "", errors.New("platform not supported")
	}
}

func ProcessorToCMakeName(platform system.Platform, cpu system.Processor) (string, error) {
	switch platform {
	case system.PlatformLinux, system.PlatformFreeBSD:
		switch cpu {
		case system.ProcessorX86:
			if platform == system.PlatformFreeBSD {
				return "i386", nil
			}
			return "i686", nil
		case system.ProcessorX64:
			if platform == system.PlatformFreeBSD {
				return "amd64", nil
			}
			return "x86_64", nil
		case system.ProcessorArm32:
			return "armv7l", nil
		case system.ProcessorArm64:
			return "aarch64", nil
		case system.ProcessorRISCV32:
			return "riscv32", nil
		case system.ProcessorRISCV64:
			return "riscv64", nil
		}
	case system.PlatformMac:
		switch cpu {
		case system.ProcessorX64:
			return "x86_64", nil
		case system.ProcessorArm64:
			return "arm64", nil
		}
	case system.PlatformWindows:
		switch cpu {
		case system.ProcessorX86:
			return "x86", nil
		case system.ProcessorX64:
			return "AMD64", nil
		case system.ProcessorArm32:
			return "ARM", nil
		case system.ProcessorArm64:
			return "ARM64", nil
		}
	}
	return "", fmt.Errorf("unsupported platform/processor combination: %s/%v", platform, cpu)
}

func GenerateToolchainFile(ctx context.Context, opts GenerateToolchainFileOptions) error {
	absWorkspaceDir, err := filepath.Abs(opts.WorkspaceDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace directory: %w", err)
	}

	cCompiler := opts.CCompiler
	cxxCompiler := opts.CXXCompiler
	linker := opts.Linker

	systemName, err := PlatformToCMakeName(opts.SystemPlatform)
	if err != nil {
		return fmt.Errorf("failed to get CMake platform name: %w", err)
	}

	systemProcessor, err := ProcessorToCMakeName(opts.SystemPlatform, opts.SystemProcessor)
	if err != nil {
		return fmt.Errorf("failed to get CMake processor name: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Automatically generated toolchain file\n")

	if systemName != "" {
		sb.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", systemName))
	}
	if systemProcessor != "" {
		sb.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", systemProcessor))
	}

	if cCompiler != "" {
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_COMPILER \"%s\")\n", cCompiler))
	}
	if cxxCompiler != "" {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_COMPILER \"%s\")\n", cxxCompiler))
	}
	if linker != "" {
		sb.WriteString(fmt.Sprintf("set(CMAKE_LINKER \"%s\")\n", linker))
	}

	// Remap debug symbols
	// We use -fdebug-prefix-map=OLD=NEW for GCC/Clang
	// We want to map the absolute workspace directory to something relative or just "."
	debugMapFlag := fmt.Sprintf("-fdebug-compilation-dir=. -fdebug-prefix-map=%s=.", absWorkspaceDir)
	cFlags := debugMapFlag
	cxxFlags := debugMapFlag

	if len(opts.ExtraCXXFlags) > 0 {
		cxxFlags += " " + strings.Join(opts.ExtraCXXFlags, " ")
	}

	sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_INIT \"%s\")\n", cFlags))
	sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_INIT \"%s\")\n", cxxFlags))

	//sb.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)\n")
	//sb.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)\n")
	//sb.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)\n")
	sb.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE NEVER)\n")
	//sb.WriteString("set(CMAKE_FIND_PACKAGE_PREFER_CONFIG TRUE)\n")

	// sb.WriteString("set(CMAKE_FIND_USE_SYSTEM_ENVIRONMENT_PATH OFF)\n")
	//sb.WriteString("set(CMAKE_FIND_USE_SYSTEM_PACKAGE_REGISTRY OFF)\n")
	//sb.WriteString("set(CMAKE_FIND_USE_PACKAGE_REGISTRY OFF)\n")

	err = os.MkdirAll(filepath.Dir(opts.OutputFile), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for toolchain file: %w", err)
	}

	err = os.WriteFile(opts.OutputFile, []byte(sb.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write toolchain file: %w", err)
	}

	return nil
}

type Option struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

func (o *Option) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		o.Value = value.Value
		o.Type = ""
		return nil
	}
	type Alias Option
	var aux Alias
	if err := value.Decode(&aux); err != nil {
		return err
	}
	o.Type = aux.Type
	o.Value = aux.Value
	return nil
}

func (o Option) MarshalYAML() (interface{}, error) {
	if o.Type == "" {
		return o.Value, nil
	}
	type Alias Option
	return Alias(o), nil
}
