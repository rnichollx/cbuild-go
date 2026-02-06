package ccommon

import (
	"path/filepath"
	"testing"
)

func TestGetSourcePathResolvesRelativeLocalAgainstWorkspace(t *testing.T) {
	ws := &WorkspaceContext{
		WorkspacePath: filepath.Join("tmp", "workspace"),
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: ".."},
			},
		},
	}

	got, err := ws.GetSourcePath("main")
	if err != nil {
		t.Fatalf("GetSourcePath returned error: %v", err)
	}

	want := filepath.Join("tmp", "workspace", "..")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGetSourcePathReturnsAbsoluteLocalAsIs(t *testing.T) {
	absLocal := filepath.Join(string(filepath.Separator), "tmp", "source")
	ws := &WorkspaceContext{
		WorkspacePath: filepath.Join("tmp", "workspace"),
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: absLocal},
			},
		},
	}

	got, err := ws.GetSourcePath("main")
	if err != nil {
		t.Fatalf("GetSourcePath returned error: %v", err)
	}

	if got != absLocal {
		t.Fatalf("expected %q, got %q", absLocal, got)
	}
}

func TestGetSourcePathErrorsWhenSourceMissing(t *testing.T) {
	ws := &WorkspaceContext{
		WorkspacePath: "workspace",
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{},
		},
	}

	_, err := ws.GetSourcePath("missing")
	if err == nil {
		t.Fatalf("expected missing source error")
	}
}
