package ccommon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaultsAddsExistingSuggestedDependency(t *testing.T) {
	tmpDir := t.TempDir()
	mainSourceDir := filepath.Join(tmpDir, "main-source")
	depSourceDir := filepath.Join(tmpDir, "dep-source")

	if err := os.MkdirAll(mainSourceDir, 0755); err != nil {
		t.Fatalf("failed to create main source dir: %v", err)
	}
	if err := os.MkdirAll(depSourceDir, 0755); err != nil {
		t.Fatalf("failed to create dep source dir: %v", err)
	}

	csetupContent := `suggested_dep_sources:
  dep:
    git:
      repository: "https://example.com/dep.git"
default_configuration: {}
`
	if err := os.WriteFile(filepath.Join(mainSourceDir, "csetup.yml"), []byte(csetupContent), 0644); err != nil {
		t.Fatalf("failed to write csetup.yml: %v", err)
	}

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: mainSourceDir},
				"dep":  {Local: depSourceDir},
			},
			Targets: map[string]*TargetConfiguration{
				"main": {Source: "main"},
				"dep":  {Source: "dep"},
			},
		},
	}

	if err := ws.LoadDefaults(context.Background(), "main"); err != nil {
		t.Fatalf("LoadDefaults returned error: %v", err)
	}

	mainTarget := ws.Config.Targets["main"]
	if mainTarget == nil {
		t.Fatalf("main target missing after LoadDefaults")
	}

	depCount := 0
	for _, dep := range mainTarget.Depends {
		if dep == "dep" {
			depCount++
		}
	}

	if depCount != 1 {
		t.Fatalf("expected dependency 'dep' exactly once, got %d (depends=%v)", depCount, mainTarget.Depends)
	}
}

func TestLoadDefaultsErrorsOnUnknownCSetupField(t *testing.T) {
	tmpDir := t.TempDir()
	mainSourceDir := filepath.Join(tmpDir, "main-source")

	if err := os.MkdirAll(mainSourceDir, 0755); err != nil {
		t.Fatalf("failed to create main source dir: %v", err)
	}

	csetupContent := `default_configuration:
  project_type: CMake
  cmake_project_name: WrongKey
`
	if err := os.WriteFile(filepath.Join(mainSourceDir, "csetup.yml"), []byte(csetupContent), 0644); err != nil {
		t.Fatalf("failed to write csetup.yml: %v", err)
	}

	ws := &WorkspaceContext{
		WorkspacePath: tmpDir,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: mainSourceDir},
			},
			Targets: map[string]*TargetConfiguration{
				"main": {Source: "main"},
			},
		},
	}

	err := ws.LoadDefaults(context.Background(), "main")
	if err == nil {
		t.Fatalf("expected LoadDefaults to fail on unknown field")
	}

	if !strings.Contains(err.Error(), "cmake_project_name") {
		t.Fatalf("expected error to mention unknown field, got: %v", err)
	}
}
