package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/host"

	"gopkg.in/yaml.v3"
)

func handleDetectToolchains(ctx context.Context, workspacePath string, args []string) error {
	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	hostOS := host.DetectHostPlatform()
	hostProcessor := host.DetectHostProcessor()
	hostKey := fmt.Sprintf("host-%s-%s", hostOS.StringLower(), hostProcessor.StringLower())

	targetSystem := hostOS
	targetArch := hostProcessor

	detectors := []struct {
		name          string
		cCompiler     string
		cxxCompiler   string
		extraCXXFlags []string
	}{
		{"system-gcc", "gcc", "g++", nil},
		{"system-clang", "clang", "clang++", nil},
		{"system-clang-libcxx", "clang", "clang++", []string{"-stdlib=libc++"}},
		{"system-gcc-libcxx", "gcc", "g++", []string{"-stdlib=libc++"}},
	}

	toolchainsDir := filepath.Join(workspacePath, "toolchains")
	err = os.MkdirAll(toolchainsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create toolchains directory: %w", err)
	}

	for _, d := range detectors {
		_, errC := exec.LookPath(d.cCompiler)
		_, errCXX := exec.LookPath(d.cxxCompiler)

		if errC == nil && errCXX == nil {
			// Check if compilers actually work by building a minimal CMake project
			testDir, err := os.MkdirTemp("", "csetup_detect_test")
			if err != nil {
				return fmt.Errorf("failed to create temp dir: %w", err)
			}
			defer os.RemoveAll(testDir)

			err = os.MkdirAll(filepath.Join(testDir, "build"), 0755)
			if err != nil {
				return fmt.Errorf("failed to create build dir: %w", err)
			}

			err = os.WriteFile(filepath.Join(testDir, "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.10)\nproject(test)\nadd_executable(test main.cpp)\n"), 0644)
			if err != nil {
				return fmt.Errorf("failed to create CMakeLists.txt: %w", err)
			}

			err = os.WriteFile(filepath.Join(testDir, "main.cpp"), []byte("int main() { return 0; }\n"), 0644)
			if err != nil {
				return fmt.Errorf("failed to create main.cpp: %w", err)
			}

			// Generate a temporary toolchain file for the test
			tcFilePath := filepath.Join(testDir, "toolchain.cmake")
			tc := ccommon.Toolchain{
				TargetArch:   targetArch,
				TargetSystem: targetSystem,
				CMakeToolchain: map[string]ccommon.CMakeToolchainOptions{
					hostKey: {
						Generate: &ccommon.CMakeGenerateToolchainFileOptions{
							CCompiler:     d.cCompiler,
							CXXCompiler:   d.cxxCompiler,
							ExtraCXXFlags: d.extraCXXFlags,
						},
					},
				},
			}

			// We need a workspace to call GenerateToolchainFile, but we can call cmake.GenerateToolchainFile directly
			err = ws.GenerateToolchainFile(tc.CMakeToolchain[hostKey].Generate, targetSystem, targetArch, tcFilePath)
			if err != nil {
				return fmt.Errorf("failed to generate test toolchain file: %w", err)
			}

			// Run CMake configure
			cmd := exec.Command("cmake", "-S", testDir, "-B", filepath.Join(testDir, "build"), "-G", "Ninja", "-DCMAKE_TOOLCHAIN_FILE="+tcFilePath)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Detected %s, but %s cannot build a hello world program, skipping.\n", d.cxxCompiler, d.name)
				continue
			}

			// Run CMake build
			cmd = exec.Command("cmake", "--build", filepath.Join(testDir, "build"))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Detected %s, but %s cannot build a hello world program, skipping.\n", d.cxxCompiler, d.name)
				continue
			}

			fmt.Printf("Detected %s, creating toolchain...\n", d.name)

			finalTc := ccommon.Toolchain{
				TargetArch:   targetArch,
				TargetSystem: targetSystem,
				CMakeToolchain: map[string]ccommon.CMakeToolchainOptions{
					hostKey: {
						Generate: &ccommon.CMakeGenerateToolchainFileOptions{
							CCompiler:     d.cCompiler,
							CXXCompiler:   d.cxxCompiler,
							ExtraCXXFlags: d.extraCXXFlags,
						},
					},
				},
			}

			tcDir := filepath.Join(toolchainsDir, d.name)
			err = os.MkdirAll(tcDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create toolchain directory for %s: %w", d.name, err)
			}

			yamlFile, err := yaml.Marshal(finalTc)
			if err != nil {
				return fmt.Errorf("failed to marshal toolchain %s: %w", d.name, err)
			}

			err = os.WriteFile(filepath.Join(tcDir, "toolchain.yml"), yamlFile, 0644)
			if err != nil {
				return fmt.Errorf("failed to write toolchain file for %s: %w", d.name, err)
			}
		} else {
			fmt.Printf("Compilers for %s not found (tried %s and %s), skipping.\n", d.name, d.cCompiler, d.cxxCompiler)
		}
	}

	return nil
}
