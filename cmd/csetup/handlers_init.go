package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleInit(workspacePath string, args []string) error {
	var reinit bool
	workspaceName := ""

	for _, arg := range args {
		if arg == "--reinit" {
			reinit = true
		} else if !strings.HasPrefix(arg, "-") && workspaceName == "" {
			workspaceName = arg
		}
	}

	if workspaceName == "" {
		return fmt.Errorf("usage: csetup init <workspace name> [--reinit]")
	}

	// Use workspaceName as the workspacePath for init
	targetPath := workspaceName

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("error creating workspace directory: %w", err)
	}

	workspaceConfig := filepath.Join(targetPath, "cbuild_workspace.yml")
	if _, err := os.Stat(workspaceConfig); err == nil {
		if !reinit {
			return fmt.Errorf("%s already exists. Use --reinit to overwrite", workspaceConfig)
		} else {
			// Delete toolchains, sources and buildspaces
			dirsToDelete := []string{"toolchains", "sources", "buildspaces"}
			for _, d := range dirsToDelete {
				dirPath := filepath.Join(targetPath, d)
				fmt.Printf("Cleaning %s...\n", dirPath)
				os.RemoveAll(dirPath)
			}
		}
	}

	ws := &ccommon.Workspace{
		WorkspacePath: targetPath,
		Targets:       make(map[string]*ccommon.TargetConfiguration),
		CXXVersion:    "20",
	}

	err := ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Initialized empty workspace in %s\n", targetPath)
	return nil
}
