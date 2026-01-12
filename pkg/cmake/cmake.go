package cmake

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/system"
	"gopkg.in/yaml.v3"
)

type CompilerType int

const (
	CompilerTypeUnknown CompilerType = iota
	CompilerTypeGCC
	CompilerTypeClang
	CompilerTypeMSVC
)

type LinkerType int

const (
	LinkerTypeUnknown LinkerType = iota
	LinkerTypeGNULD
	LinkerTypeLLD
)

type BuildPreset int

const (
	BuildPresetNone BuildPreset = iota
	BuildPresetDebug
	BuildPresetRelease
	BuildPresetQuick
	BuildPresetRelWithDebInfo
	BuildPresetDebugASAN
	BuildPresetDebugTSAN
	BuildPresetDebugCoverage
	BuildPresetProfile
	BuildPresetReleaseASAN
	BuildPresetReleaseTSAN
)

type GenerateToolchainFileOptions struct {
	CompilerType       CompilerType
	CCompiler          string
	CXXCompiler        string
	Linker             string
	ExtraCompilerFlags []string
	ExtraCFlags        []string
	ExtraCXXFlags      []string
	SystemPlatform     system.Platform
	SystemProcessor    system.Processor
	WorkspaceDir       string
	OutputFile         string
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

var clangRE = regexp.MustCompile("clang(\\+\\+)?(-\\d+)?$")
var gccRE = regexp.MustCompile("(gcc|g\\+\\+)(-\\d+)?$")
var msvcRE = regexp.MustCompile("(?i)^cl(\\.exe)?$")

func guessCompilerType(opts *GenerateToolchainFileOptions) CompilerType {

	if opts.CCompiler != "" {
		base := filepath.Base(opts.CCompiler)
		fmt.Printf("Generate options cc: %q\n", base)
		if clangRE.MatchString(base) {
			return CompilerTypeClang
		}
		if gccRE.MatchString(base) {
			return CompilerTypeGCC
		}
		if msvcRE.MatchString(base) {
			return CompilerTypeMSVC
		}
	}
	if opts.CXXCompiler != "" {

		base := filepath.Base(opts.CXXCompiler)
		fmt.Printf("Generate options c++: %q\n", base)
		if clangRE.MatchString(base) {
			return CompilerTypeClang
		}
		if gccRE.MatchString(base) {
			return CompilerTypeGCC
		}
		if msvcRE.MatchString(base) {
			return CompilerTypeMSVC
		}
	}
	return CompilerTypeUnknown
}

func GenerateToolchainFile(ctx context.Context, opts GenerateToolchainFileOptions) error {

	if opts.CompilerType == CompilerTypeUnknown {
		opts.CompilerType = guessCompilerType(&opts)
		fmt.Printf("Guessing compiler type: %s\n", opts.CompilerType)
		if opts.CompilerType == CompilerTypeUnknown {
			return errors.New("unknown compiler")
		}
	}

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

	var debugFlags []string
	var ASANFlags []string
	var TSANFlags []string
	var releaseFlags []string
	var quickFlags []string
	var profileFlags []string
	var debugCoverageFlags []string
	var debugASANFlags []string
	var debugTSANFlags []string
	var relASANFlags []string
	var relTSANFlags []string

	supportedConfigs := []string{"Debug", "Release", "RelWithDebInfo", "Quick", "Profile"}

	if opts.CompilerType == CompilerTypeClang {
		debugFlags = append(debugFlags, "-fdebug-compilation-dir=.")
	}
	if opts.CompilerType == CompilerTypeClang || opts.CompilerType == CompilerTypeGCC {

		debugFlags = append(debugFlags, fmt.Sprintf("-fdebug-prefix-map=%s=.", absWorkspaceDir))
	}

	if opts.CompilerType == CompilerTypeClang || opts.CompilerType == CompilerTypeGCC {
		debugFlags = append(debugFlags, "-g")
		debugFlags = append(debugFlags, "-Og")

		ASANFlags = append(ASANFlags, "-fsanitize=address")
		ASANFlags = append(ASANFlags, "-fsanitize=undefined")
		TSANFlags = append(TSANFlags, "-fsanitize=thread")
		TSANFlags = append(TSANFlags, "-fsanitize=undefined")

		releaseFlags = append(releaseFlags, "-O3")
		releaseFlags = append(releaseFlags, "-DNDEBUG")

		quickFlags = append(quickFlags, "-O1")
		quickFlags = append(quickFlags, "-DNDEBUG")

		profileFlags = append(profileFlags, "-O3")
		profileFlags = append(profileFlags, "-g")
		profileFlags = append(profileFlags, "-DNDEBUG")

		debugCoverageFlags = append(debugCoverageFlags, debugFlags...)
		debugCoverageFlags = append(debugCoverageFlags, "--coverage")

		debugASANFlags = append(debugASANFlags, debugFlags...)
		debugASANFlags = append(debugASANFlags, ASANFlags...)

		debugTSANFlags = append(debugTSANFlags, debugFlags...)
		debugTSANFlags = append(debugTSANFlags, TSANFlags...)

		relASANFlags = append(relASANFlags, releaseFlags...)
		relASANFlags = append(relASANFlags, ASANFlags...)

		relTSANFlags = append(relTSANFlags, releaseFlags...)
		relTSANFlags = append(relTSANFlags, TSANFlags...)

		supportedConfigs = append(supportedConfigs, "DebugCoverage", "DebugASAN", "DebugTSAN", "ReleaseASAN", "ReleaseTSAN")
	}

	if opts.CompilerType == CompilerTypeMSVC {
		debugFlags = append(debugFlags, "/Zi")
		debugFlags = append(debugFlags, "/Od")
		debugFlags = append(debugFlags, "/RTC1")

		releaseFlags = append(releaseFlags, "/O2")
		releaseFlags = append(releaseFlags, "/DNDEBUG")

		quickFlags = append(quickFlags, "/O1")
		quickFlags = append(quickFlags, "/DNDEBUG")

		profileFlags = append(profileFlags, "/O2")
		profileFlags = append(profileFlags, "/Zi")
		profileFlags = append(profileFlags, "/DNDEBUG")

		// MSVC support for ASAN is limited/different, leaving for now
		debugASANFlags = append(debugASANFlags, debugFlags...)
		debugTSANFlags = append(debugTSANFlags, debugFlags...)
		relASANFlags = append(relASANFlags, releaseFlags...)
		relTSANFlags = append(relTSANFlags, releaseFlags...)
		debugCoverageFlags = append(debugCoverageFlags, debugFlags...)
	}

	var commonFlags []string
	commonFlags = append(commonFlags, opts.ExtraCompilerFlags...)

	cFlags := append(commonFlags, opts.ExtraCFlags...)
	cxxFlags := append(commonFlags, opts.ExtraCXXFlags...)

	sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_INIT %q)\n", strings.Join(cFlags, " ")))
	sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_INIT %q)\n", strings.Join(cxxFlags, " ")))

	if len(supportedConfigs) > 0 {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CONFIGURATION_TYPES %q CACHE STRING \"\" FORCE)\n", strings.Join(supportedConfigs, ";")))
	}

	// Debug
	if contains(supportedConfigs, "Debug") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_DEBUG_INIT %q)\n", strings.Join(debugFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_DEBUG_INIT %q)\n", strings.Join(debugFlags, " ")))
	}

	// Release
	if contains(supportedConfigs, "Release") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_RELEASE_INIT %q)\n", strings.Join(releaseFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_RELEASE_INIT %q)\n", strings.Join(releaseFlags, " ")))
	}

	// RelWithDebInfo
	if contains(supportedConfigs, "RelWithDebInfo") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_RELWITHDEBINFO_INIT %q)\n", strings.Join(profileFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_RELWITHDEBINFO_INIT %q)\n", strings.Join(profileFlags, " ")))
	}

	// Quick
	if contains(supportedConfigs, "Quick") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_QUICK_INIT %q)\n", strings.Join(quickFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_QUICK_INIT %q)\n", strings.Join(quickFlags, " ")))
	}

	// Profile
	if contains(supportedConfigs, "Profile") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_PROFILE_INIT %q)\n", strings.Join(profileFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_PROFILE_INIT %q)\n", strings.Join(profileFlags, " ")))
	}

	// DebugCoverage
	if contains(supportedConfigs, "DebugCoverage") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_DEBUGCOVERAGE_INIT %q)\n", strings.Join(debugCoverageFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_DEBUGCOVERAGE_INIT %q)\n", strings.Join(debugCoverageFlags, " ")))
	}

	// DebugASAN
	if contains(supportedConfigs, "DebugASAN") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_DEBUGASAN_INIT %q)\n", strings.Join(debugASANFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_DEBUGASAN_INIT %q)\n", strings.Join(debugASANFlags, " ")))
	}

	// DebugTSAN
	if contains(supportedConfigs, "DebugTSAN") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_DEBUGTSAN_INIT %q)\n", strings.Join(debugTSANFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_DEBUGTSAN_INIT %q)\n", strings.Join(debugTSANFlags, " ")))
	}

	// ReleaseASAN
	if contains(supportedConfigs, "ReleaseASAN") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_RELEASEASAN_INIT %q)\n", strings.Join(relASANFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_RELEASEASAN_INIT %q)\n", strings.Join(relASANFlags, " ")))
	}

	// ReleaseTSAN
	if contains(supportedConfigs, "ReleaseTSAN") {
		sb.WriteString(fmt.Sprintf("set(CMAKE_CXX_FLAGS_RELEASETSAN_INIT %q)\n", strings.Join(relTSANFlags, " ")))
		sb.WriteString(fmt.Sprintf("set(CMAKE_C_FLAGS_RELEASETSAN_INIT %q)\n", strings.Join(relTSANFlags, " ")))
	}

	sb.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE NEVER)\n")

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

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
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
