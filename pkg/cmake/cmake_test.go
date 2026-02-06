package cmake

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/rpnx/cbuild-go/pkg/system"
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

func TestPlatformToCMakeName(t *testing.T) {
	tests := []struct {
		platform system.Platform
		want     string
		wantErr  bool
	}{
		{system.PlatformMac, "Darwin", false},
		{system.PlatformLinux, "Linux", false},
		{system.PlatformFreeBSD, "FreeBSD", false},
		{system.PlatformOpenBSD, "OpenBSD", false},
		{system.PlatformWindows, "Windows", false},
		{system.PlatformUnknown, "", true},
	}

	for _, tt := range tests {
		got, err := PlatformToCMakeName(tt.platform)
		if (err != nil) != tt.wantErr {
			t.Errorf("PlatformToCMakeName(%v) error = %v, wantErr %v", tt.platform, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("PlatformToCMakeName(%v) = %v, want %v", tt.platform, got, tt.want)
		}
	}
}

func TestGenerateToolchainFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "toolchain.cmake")

	opts := GenerateToolchainFileOptions{
		CompilerType:       CompilerTypeGCC,
		CCompiler:          "gcc",
		CXXCompiler:        "g++",
		ExtraCompilerFlags: []string{"-O3"},
		ExtraCFlags:        []string{"-std=c11"},
		ExtraCXXFlags:      []string{"-std=c++17"},
		SystemPlatform:     system.PlatformLinux,
		SystemProcessor:    system.ProcessorX64,
		WorkspaceDir:       ".",
		OutputFile:         outputFile,
	}

	err := GenerateToolchainFile(nil, opts)
	if err != nil {
		t.Fatalf("GenerateToolchainFile failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	sContent := string(content)
	expected := []string{
		"set(CMAKE_C_COMPILER \"gcc\")",
		"set(CMAKE_CXX_COMPILER \"g++\")",
		"set(CMAKE_C_FLAGS_INIT \"-O3 -std=c11\")",
		"set(CMAKE_CONFIGURATION_TYPES \"Debug;Release;RelWithDebInfo;Quick;Profile;DebugCoverage;DebugASAN;DebugTSAN;ReleaseASAN;ReleaseTSAN\" CACHE STRING \"\" FORCE)",
		"set(CMAKE_C_FLAGS_QUICK_INIT \"-O1 -DNDEBUG\")",
		"-fsanitize=address",
	}

	for _, e := range expected {
		if !strings.Contains(sContent, e) {
			t.Errorf("Expected content %q not found in toolchain file:\n%s", e, sContent)
		}
	}
}
