package ccommon

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"gitlab.com/rpnx/cbuild-go/pkg/cmake"
)

func TestCMakeConfigureArgsForcesCMakeCxxExtensionsOff(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false, false)

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: "src/main"},
			},
		},
	}

	target := &TargetContext{
		Name: "main",
		Config: TargetConfiguration{
			ExtraCMakeConfigureArgs: []string{
				"-DCMAKE_CXX_EXTENSIONS=ON",
			},
			CMakeOptions: map[string]cmake.Option{
				"CMAKE_CXX_EXTENSIONS": {
					Value: "ON",
				},
			},
		},
	}

	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	if len(args) == 0 {
		t.Fatalf("expected configure args to be non-empty")
	}

	lastArg := args[len(args)-1]
	if lastArg != "-DCMAKE_CXX_EXTENSIONS=OFF" {
		t.Fatalf("expected last arg to force CXX extensions OFF, got %q", lastArg)
	}

	hasOnArg := false
	for _, arg := range args {
		if arg == "-DCMAKE_CXX_EXTENSIONS=ON" || arg == "-DCMAKE_CXX_EXTENSIONS:STRING=ON" {
			hasOnArg = true
			break
		}
	}
	if !hasOnArg {
		t.Fatalf("expected test setup to include a conflicting CMAKE_CXX_EXTENSIONS=ON arg; got args: %v", args)
	}
}

func TestCMakeConfigureArgsIncludesCMakeCxxExtensionsOff(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false, false)

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: "src/main"},
			},
			CXXVersion: "20",
		},
	}

	target := &TargetContext{
		Name: "main",
		Config: TargetConfiguration{
			CxxStandard: stringPtr("17"),
		},
	}

	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	required := []string{
		"-S",
		filepath.Join(tmpDir, "src/main"),
		"-B",
		filepath.Join(tmpDir, "buildspaces", "tc", "main", "Debug"),
		"-G",
		"Ninja",
		"-DCMAKE_BUILD_TYPE=Debug",
		"-DCMAKE_EXPORT_COMPILE_COMMANDS=ON",
		"-DCMAKE_CXX_STANDARD=17",
		"-DBUILD_TESTING=ON",
	}

	for _, want := range required {
		if !containsString(args, want) {
			t.Fatalf("expected arg %q in configure args, got: %v", want, args)
		}
	}

	if !containsString(args, "-DCMAKE_CXX_EXTENSIONS=OFF") {
		t.Fatalf("expected CMAKE_CXX_EXTENSIONS=OFF in configure args, got: %v", args)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func stringPtr(value string) *string {
	return &value
}

func TestCMakeConfigureArgsWithAbsoluteSourcePath(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false, false)

	absSource := filepath.Join(tmpDir, "abs-src")
	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: absSource},
			},
		},
	}

	target := &TargetContext{Name: "main"}
	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Release",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	if !containsString(args, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", "Release")) {
		t.Fatalf("missing build type arg in: %v", args)
	}
	if !containsString(args, "-DCMAKE_CXX_EXTENSIONS=OFF") {
		t.Fatalf("expected CMAKE_CXX_EXTENSIONS=OFF in configure args, got: %v", args)
	}
}

func TestCMakeConfigureArgsIncludesTransitiveDependencyDirs(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false, false)

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"app":  {Local: "src/app"},
				"mid":  {Local: "src/mid"},
				"leaf": {Local: "src/leaf"},
			},
			Targets: map[string]*TargetConfiguration{
				"app": {
					Depends: []string{"mid"},
				},
				"mid": {
					Depends: []string{"leaf"},
				},
				"leaf": {},
			},
		},
	}

	target := &TargetContext{
		Name:   "app",
		Config: *ws.Config.Targets["app"],
	}

	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	midConfigPath := filepath.Join(tmpDir, "buildspaces", "tc", "mid", "Debug")
	leafConfigPath := filepath.Join(tmpDir, "buildspaces", "tc", "leaf", "Debug")

	if !containsString(args, fmt.Sprintf("-Dmid_DIR=%s", midConfigPath)) {
		t.Fatalf("expected transitive dependency arg for mid, got: %v", args)
	}
	if !containsString(args, fmt.Sprintf("-Dleaf_DIR=%s", leafConfigPath)) {
		t.Fatalf("expected transitive dependency arg for leaf, got: %v", args)
	}
}

func TestCMakeConfigureArgsIncludesTransitiveStagedDependencyPaths(t *testing.T) {
	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false, false)

	staged := true
	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"app":  {Local: "src/app"},
				"mid":  {Local: "src/mid"},
				"leaf": {Local: "src/leaf"},
			},
			Targets: map[string]*TargetConfiguration{
				"app": {
					Depends: []string{"mid"},
				},
				"mid": {
					Depends: []string{"leaf"},
				},
				"leaf": {
					Staged: &staged,
				},
			},
		},
	}

	target := &TargetContext{
		Name:   "app",
		Config: *ws.Config.Targets["app"],
	}

	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	leafStagingPath := filepath.Join(tmpDir, "staging", "tc", "Debug", "leaf")

	if !containsString(args, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", leafStagingPath)) {
		t.Fatalf("expected transitive staged dependency prefix path, got: %v", args)
	}
	if !containsString(args, fmt.Sprintf("-DCMAKE_MODULE_PATH=%s", leafStagingPath)) {
		t.Fatalf("expected transitive staged dependency module path, got: %v", args)
	}
}
