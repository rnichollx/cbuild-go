package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleInit(workspacePath string, args []string) {
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
		fmt.Println("Usage: csetup init <workspace name> [--reinit]")
		os.Exit(1)
	}

	// Use workspaceName as the workspacePath for init
	targetPath := workspaceName

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating workspace directory: %v\n", err)
		os.Exit(1)
	}

	workspaceConfig := filepath.Join(targetPath, "cbuild_workspace.yml")
	if _, err := os.Stat(workspaceConfig); err == nil {
		if !reinit {
			fmt.Fprintf(os.Stderr, "Error: %s already exists. Use --reinit to overwrite.\n", workspaceConfig)
			os.Exit(1)
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
		Targets:       make(map[string]*ccommon.Target),
		CXXVersion:    "20",
	}

	err := ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized empty workspace in %s\n", targetPath)
}
