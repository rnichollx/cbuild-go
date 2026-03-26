package ccommon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDropSourceFilesDeletesFolderInsideSources(t *testing.T) {
	workspace := t.TempDir()
	sourceDir := filepath.Join(workspace, "sources", "main")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	ws := &WorkspaceContext{
		WorkspacePath: workspace,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Git: &GitSource{Repository: "https://example.com/repo.git"}},
			},
		},
	}

	if err := ws.DropSourceFiles(context.Background(), "main"); err != nil {
		t.Fatalf("DropSourceFiles returned error: %v", err)
	}

	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Fatalf("expected source directory to be removed, stat error: %v", err)
	}
}

func TestDropSourceFilesIgnoresNonGitSource(t *testing.T) {
	workspace := t.TempDir()
	outsideDir := filepath.Join(workspace, "outside")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("failed to create outside directory: %v", err)
	}

	ws := &WorkspaceContext{
		WorkspacePath: workspace,
		Config: WorkspaceConfig{
			Sources: map[string]*CodeSource{
				"main": {Local: "../outside"},
			},
		},
	}

	if err := ws.DropSourceFiles(context.Background(), "main"); err != nil {
		t.Fatalf("DropSourceFiles returned error for non-git source: %v", err)
	}

	if _, statErr := os.Stat(outsideDir); statErr != nil {
		t.Fatalf("outside directory should not be deleted, stat error: %v", statErr)
	}
}
