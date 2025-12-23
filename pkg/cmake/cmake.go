package cmake

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GenerateToolchainFileOptions struct {
	CCompiler       string
	CXXCompiler     string
	Linker          string
	ExtraCXXFlags   []string
	SystemName      string
	SystemProcessor string
	WorkspaceDir    string
	OutputFile      string
}

func GenerateToolchainFile(opts GenerateToolchainFileOptions) error {
	absWorkspaceDir, err := filepath.Abs(opts.WorkspaceDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace directory: %w", err)
	}

	cCompiler := opts.CCompiler
	cxxCompiler := opts.CXXCompiler
	linker := opts.Linker

	systemProcessor := opts.SystemProcessor
	switch strings.ToLower(systemProcessor) {
	case "x64":
		systemProcessor = "x86_64"
	case "x86":
		systemProcessor = "i386"
	case "arm32":
		systemProcessor = "arm"
	}

	systemName := opts.SystemName
	switch strings.ToLower(systemName) {
	case "linux":
		systemName = "Linux"
	case "windows":
		systemName = "Windows"
	case "macos":
		systemName = "Darwin"
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
	debugMapFlag := fmt.Sprintf("-fdebug-prefix-map=%s=.", absWorkspaceDir)
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
