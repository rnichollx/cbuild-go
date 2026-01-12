package ccommon

import (
	"gitlab.com/rpnx/cbuild-go/pkg/cmake"
	"gitlab.com/rpnx/cbuild-go/pkg/system"
)

type CMakeToolchainOptions struct {
	CMakeToolchainFile string                             `yaml:"cmake_toolchain_file,omitempty"`
	Generate           *CMakeGenerateToolchainFileOptions `yaml:"generate,omitempty"`
}

type CMakeGenerateToolchainFileOptions struct {
	CompilerType       cmake.CompilerType `yaml:"compiler_type,omitempty"`
	CCompiler          string             `yaml:"c_compiler"`
	CXXCompiler        string             `yaml:"cxx_compiler"`
	Linker             string             `yaml:"linker,omitempty"`
	ExtraCompilerFlags []string           `yaml:"extra_compiler_flags,omitempty"`
	ExtraCXXFlags      []string           `yaml:"extra_cxx_flags,omitempty"`
	ExtraCFlags        []string           `yaml:"extra_c_flags,omitempty"`
}

type Toolchain struct {
	CMakeToolchain map[string]CMakeToolchainOptions `yaml:"cmake_toolchain"`
	TargetArch     system.Processor                 `yaml:"target_arch"`
	TargetSystem   system.Platform                  `yaml:"target_system"`
}

type TargetBuildParameters struct {
	Toolchain string
	BuildType string
	DryRun    bool
}
