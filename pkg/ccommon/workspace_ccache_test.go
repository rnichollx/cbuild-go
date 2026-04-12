package ccommon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCCacheEnvironmentUsesPerToolchainWorkspaceCacheDir(t *testing.T) {
	tmpDir := t.TempDir()

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
	}

	env, err := ws.CCacheEnvironment("tc", true)
	if err != nil {
		t.Fatalf("CCacheEnvironment failed: %v", err)
	}

	expectedDir := filepath.Join(tmpDir, "cache", "tc", "ccache")
	expectedDir, err = filepath.Abs(expectedDir)
	if err != nil {
		t.Fatalf("failed to resolve expected ccache dir: %v", err)
	}

	if !containsEnvEntry(env, "CCACHE_DIR="+expectedDir) {
		t.Fatalf("expected CCACHE_DIR=%s in env: %v", expectedDir, env)
	}
	if !containsEnvEntry(env, "CCACHE_COMPILERCHECK=content") {
		t.Fatalf("expected CCACHE_COMPILERCHECK=content in env: %v", env)
	}
	if !containsEnvEntry(env, "CCACHE_NODIRECT=1") {
		t.Fatalf("expected CCACHE_NODIRECT=1 in env: %v", env)
	}

	if !isDir(expectedDir) {
		t.Fatalf("expected ccache dir to be created at %s", expectedDir)
	}
}

func TestCCacheEnvironmentSkipsSettingsWhenDisabled(t *testing.T) {
	ws := &WorkspaceContext{
		WorkspacePath: t.TempDir(),
	}

	env, err := ws.CCacheEnvironment("tc", false)
	if err != nil {
		t.Fatalf("CCacheEnvironment failed: %v", err)
	}

	for _, entry := range env {
		if entry == "CCACHE_COMPILERCHECK=content" || entry == "CCACHE_NODIRECT=1" {
			t.Fatalf("did not expect ccache setting in env when disabled: %v", env)
		}
	}
}

func TestCMakeConfigureArgsAddsCCacheLauncherWhenToolchainEnablesIt(t *testing.T) {
	oldLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		if file == "ccache" {
			return "/usr/bin/ccache", nil
		}
		return oldLookPath(file)
	}
	t.Cleanup(func() {
		execLookPath = oldLookPath
	})

	tmpDir := t.TempDir()
	writeGeneratedToolchain(t, tmpDir, "tc", false, true)

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: "src/main"},
			},
		},
	}

	target := &TargetContext{Name: "main"}
	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	if !containsString(args, "-DCMAKE_C_COMPILER_LAUNCHER=/usr/bin/ccache") {
		t.Fatalf("expected C compiler launcher in configure args, got: %v", args)
	}
	if !containsString(args, "-DCMAKE_CXX_COMPILER_LAUNCHER=/usr/bin/ccache") {
		t.Fatalf("expected CXX compiler launcher in configure args, got: %v", args)
	}
}

func TestCMakeConfigureArgsSkipsCCacheLauncherWhenToolchainDisablesIt(t *testing.T) {
	oldLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		if file == "ccache" {
			return "/usr/bin/ccache", nil
		}
		return oldLookPath(file)
	}
	t.Cleanup(func() {
		execLookPath = oldLookPath
	})

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

	target := &TargetContext{Name: "main"}
	args, err := target.CMakeConfigureArgs(context.Background(), ws, TargetBuildParameters{
		Toolchain: "tc",
		BuildType: "Debug",
	})
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	if containsString(args, "-DCMAKE_C_COMPILER_LAUNCHER=/usr/bin/ccache") {
		t.Fatalf("did not expect C compiler launcher in configure args, got: %v", args)
	}
	if containsString(args, "-DCMAKE_CXX_COMPILER_LAUNCHER=/usr/bin/ccache") {
		t.Fatalf("did not expect CXX compiler launcher in configure args, got: %v", args)
	}
}

func containsEnvEntry(entries []string, want string) bool {
	for _, entry := range entries {
		if entry == want {
			return true
		}
	}
	return false
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
