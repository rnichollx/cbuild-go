package ccommon

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
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

func TestLoadDefaultsProjectDefaultsAreIgnoredByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	mainSourceDir := filepath.Join(tmpDir, "main-source")
	if err := os.MkdirAll(mainSourceDir, 0755); err != nil {
		t.Fatalf("failed to create main source dir: %v", err)
	}

	csetupContent := `default_configuration: {}
project_default_configuration:
  cxx_version: "23"
  default_build_configurations: ["DebugASAN"]
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
			Configurations: []string{"Debug", "Release"},
		},
	}

	if err := ws.LoadDefaults(context.Background(), "main"); err != nil {
		t.Fatalf("LoadDefaults returned error: %v", err)
	}

	if ws.Config.CXXVersion != "" {
		t.Fatalf("expected global cxx_version to remain unset by default, got %q", ws.Config.CXXVersion)
	}
	if !reflect.DeepEqual(ws.Config.Configurations, []string{"Debug", "Release"}) {
		t.Fatalf("expected configurations to remain unchanged by default, got %v", ws.Config.Configurations)
	}
}

func TestLoadDefaultsProjectDefaultsAppliedWhenEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	mainSourceDir := filepath.Join(tmpDir, "main-source")
	if err := os.MkdirAll(mainSourceDir, 0755); err != nil {
		t.Fatalf("failed to create main source dir: %v", err)
	}

	csetupContent := `default_configuration: {}
project_default_configuration:
  cxx_version: "23"
  default_build_configurations: ["DebugASAN", "ReleaseASAN"]
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
			Configurations: []string{"Debug", "Release"},
		},
	}

	ctx := context.Background()
	ctx = WithApplyProjectDefaults(ctx, true)

	if err := ws.LoadDefaults(ctx, "main"); err != nil {
		t.Fatalf("LoadDefaults returned error: %v", err)
	}

	if ws.Config.CXXVersion != "23" {
		t.Fatalf("expected global cxx_version to be set from project defaults, got %q", ws.Config.CXXVersion)
	}
	if !reflect.DeepEqual(ws.Config.Configurations, []string{"DebugASAN", "ReleaseASAN"}) {
		t.Fatalf("expected configurations to be set from project defaults, got %v", ws.Config.Configurations)
	}
}

func TestLoadDefaultsProjectDefaultsDoNotPropagateToDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	mainSourceDir := filepath.Join(tmpDir, "main-source")
	depSourceDir := filepath.Join(tmpDir, "sources", "dep")
	if err := os.MkdirAll(mainSourceDir, 0755); err != nil {
		t.Fatalf("failed to create main source dir: %v", err)
	}
	if err := os.MkdirAll(depSourceDir, 0755); err != nil {
		t.Fatalf("failed to create dep source dir: %v", err)
	}

	mainCsetup := `suggested_dep_sources:
  dep:
    git:
      repository: "https://example.com/dep.git"
default_configuration: {}
project_default_configuration:
  cxx_version: "20"
  default_build_configurations: ["Quick"]
`
	if err := os.WriteFile(filepath.Join(mainSourceDir, "csetup.yml"), []byte(mainCsetup), 0644); err != nil {
		t.Fatalf("failed to write main csetup.yml: %v", err)
	}

	depCsetup := `default_configuration: {}
project_default_configuration:
  cxx_version: "17"
  default_build_configurations: ["DebugTSAN"]
`
	if err := os.WriteFile(filepath.Join(depSourceDir, "csetup.yml"), []byte(depCsetup), 0644); err != nil {
		t.Fatalf("failed to write dep csetup.yml: %v", err)
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
			Configurations: []string{"Debug", "Release"},
		},
	}

	ctx := context.Background()
	var err error
	ctx, err = cli.SetBool(ctx, DownloadDepsParameter, true)
	if err != nil {
		t.Fatalf("failed to set download-deps: %v", err)
	}
	ctx = WithApplyProjectDefaults(ctx, true)

	if err := ws.LoadDefaults(ctx, "main"); err != nil {
		t.Fatalf("LoadDefaults returned error: %v", err)
	}

	if ws.Config.CXXVersion != "20" {
		t.Fatalf("expected root project defaults to remain applied, got %q", ws.Config.CXXVersion)
	}
	if !reflect.DeepEqual(ws.Config.Configurations, []string{"Quick"}) {
		t.Fatalf("expected dependency project defaults not to override root defaults, got %v", ws.Config.Configurations)
	}
}
