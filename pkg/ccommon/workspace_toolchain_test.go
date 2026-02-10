package ccommon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/rpnx/cbuild-go/pkg/host"
)

func writeGeneratedToolchain(t *testing.T, workspaceDir string, toolchainName string, separateByBuildConfig bool) {
	t.Helper()

	hostKey := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform().StringLower(), host.DetectHostProcessor().StringLower())
	separateLine := ""
	if separateByBuildConfig {
		separateLine = "      toolchain_per_buildtype: true\n"
	}

	content := fmt.Sprintf(`target_arch: "x64"
target_system: "linux"
cmake_toolchain:
  %s:
    generate:
      c_compiler: "gcc"
      cxx_compiler: "g++"
%s`, hostKey, separateLine)

	toolchainDir := filepath.Join(workspaceDir, "toolchains", toolchainName)
	if err := os.MkdirAll(toolchainDir, 0755); err != nil {
		t.Fatalf("failed to create toolchain dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolchainDir, "toolchain.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write toolchain.yml: %v", err)
	}
}

func TestToolchainFilePathGeneratedSharedByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false)

	ws := &WorkspaceContext{WorkspacePath: tmpDir}
	p, err := ws.ToolchainFilePath(context.Background(), nil, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("ToolchainFilePath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buildspaces", "tc", "generated_toolchain.cmake")
	expected, err = filepath.Abs(expected)
	if err != nil {
		t.Fatalf("failed to resolve expected path: %v", err)
	}

	if p != expected {
		t.Fatalf("expected %q, got %q", expected, p)
	}
}

func TestToolchainFilePathGeneratedPerBuildConfig(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", true)

	ws := &WorkspaceContext{WorkspacePath: tmpDir}
	p, err := ws.ToolchainFilePath(context.Background(), nil, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("ToolchainFilePath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buildspaces", "tc", "generated-toolchain-Debug.cmake")
	expected, err = filepath.Abs(expected)
	if err != nil {
		t.Fatalf("failed to resolve expected path: %v", err)
	}

	if p != expected {
		t.Fatalf("expected %q, got %q", expected, p)
	}
}

func TestPrebuildGeneratesSeparateToolchainPerBuildConfig(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", true)

	ws := &WorkspaceContext{WorkspacePath: tmpDir}

	debugPath, err := ws.Prebuild(context.Background(), TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("Prebuild(Debug) failed: %v", err)
	}

	releasePath, err := ws.Prebuild(context.Background(), TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Release",
	})
	if err != nil {
		t.Fatalf("Prebuild(Release) failed: %v", err)
	}

	if debugPath == releasePath {
		t.Fatalf("expected different toolchain files per build config, got one path: %q", debugPath)
	}

	debugContent, err := os.ReadFile(debugPath)
	if err != nil {
		t.Fatalf("failed to read debug toolchain: %v", err)
	}
	if !strings.Contains(string(debugContent), "set(CMAKE_C_FLAGS_DEBUG_INIT") {
		t.Fatalf("expected debug toolchain to include Debug flags:\n%s", string(debugContent))
	}
	if strings.Contains(string(debugContent), "set(CMAKE_C_FLAGS_RELEASE_INIT") {
		t.Fatalf("expected debug toolchain to exclude Release flags:\n%s", string(debugContent))
	}

	releaseContent, err := os.ReadFile(releasePath)
	if err != nil {
		t.Fatalf("failed to read release toolchain: %v", err)
	}
	if !strings.Contains(string(releaseContent), "set(CMAKE_C_FLAGS_RELEASE_INIT") {
		t.Fatalf("expected release toolchain to include Release flags:\n%s", string(releaseContent))
	}
	if strings.Contains(string(releaseContent), "set(CMAKE_C_FLAGS_DEBUG_INIT") {
		t.Fatalf("expected release toolchain to exclude Debug flags:\n%s", string(releaseContent))
	}
}

func TestToolchainFilePathPerBuildConfigRequiresConfig(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", true)

	ws := &WorkspaceContext{WorkspacePath: tmpDir}
	_, err := ws.ToolchainFilePath(context.Background(), nil, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "",
	})
	if err == nil {
		t.Fatalf("expected error when build config is empty for toolchain_per_buildtype")
	}
	if !strings.Contains(err.Error(), "requires a build config") {
		t.Fatalf("unexpected error: %v", err)
	}
}
