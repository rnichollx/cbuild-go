package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func handleListSources(workspacePath string, args []string) error {
	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	sourcesDir := filepath.Join(workspacePath, "sources")
	dirEntries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return fmt.Errorf("error reading sources directory: %w", err)
	}

	trackedDirs := make(map[string]bool)

	// First, list all targets from the workspace configuration
	for name, target := range ws.Targets {
		status := "[MISSING]"
		srcPath, err := target.CMakeSourcePath(ws, name)
		if err == nil {
			if info, err := os.Stat(srcPath); err == nil && info.IsDir() {
				if target.ExternalSourceOverride != nil {
					status = "[OK EXTERNAL]"
				} else {
					status = "[OK]"
				}
			}
		}

		fmt.Printf("%s %s\n", name, status)

		// If it's a standard source (in the sources/ directory), mark it as tracked
		if target.ExternalSourceOverride == nil {
			trackedDirs[name] = true
		} else {
			// If it's an override, check if it points into our sources dir anyway
			rel, err := filepath.Rel(sourcesDir, srcPath)
			if err == nil && !strings.HasPrefix(rel, "..") && rel != ".." {
				// It points inside sourcesDir, possibly to a different name
				parts := strings.Split(rel, string(os.PathSeparator))
				if len(parts) > 0 {
					trackedDirs[parts[0]] = true
				}
			}
		}
	}

	// Now list untracked folders in the sources directory
	for _, entry := range dirEntries {
		if entry.IsDir() {
			name := entry.Name()
			if !trackedDirs[name] {
				fmt.Printf("%s [UNTRACKED]\n", name)
			}
		}
	}
	return nil
}

func handleRemoveSource(workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup remove-source <source> [-D|--delete]")
	}

	source := ""
	removeFolder := false

	for _, arg := range args {
		if arg == "-D" || arg == "--delete" {
			removeFolder = true
		} else if !strings.HasPrefix(arg, "-") && source == "" {
			source = arg
		}
	}

	if source == "" {
		return fmt.Errorf("usage: csetup remove-source <source> [-D|--delete]")
	}

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	if _, ok := ws.Targets[source]; !ok {
		return fmt.Errorf("source %s not found in workspace", source)
	}

	delete(ws.Targets, source)
	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed source %s from workspace\n", source)

	if removeFolder {
		sourceDir := filepath.Join(workspacePath, "sources", source)
		if _, err := os.Stat(sourceDir); err == nil {
			fmt.Printf("Deleting source folder: %s\n", sourceDir)
			err = os.RemoveAll(sourceDir)
			if err != nil {
				return fmt.Errorf("error deleting source folder: %w", err)
			}
		} else {
			fmt.Printf("Source folder %s not found, skipping deletion.\n", sourceDir)
		}
	} else {
		fmt.Printf("Note: files in sources/%s were NOT deleted. Use -R to delete them.\n", source)
	}
	return nil
}

func handleGitClone(workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup git-clone <repo_url> [dest_name]")
	}

	repoURL := args[0]
	destName := ""
	if len(args) >= 2 {
		destName = args[1]
	} else {
		// Extract destName from repoURL
		base := filepath.Base(repoURL)
		destName = strings.TrimSuffix(base, ".git")
	}

	// 1. Check if cbuild_workspace.yml exists
	workspaceConfig := filepath.Join(workspacePath, "cbuild_workspace.yml")
	if _, err := os.Stat(workspaceConfig); os.IsNotExist(err) {
		return fmt.Errorf("%s not found. csetup must be run in a cbuild workspace", workspaceConfig)
	}

	// 2. Clone into sources/<destName>
	destDir := filepath.Join(workspacePath, "sources", destName)

	fmt.Printf("Cloning %s into %s...\n", repoURL, destDir)

	cmd := exec.Command("git", "clone", repoURL, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error cloning repository: %w", err)
	}

	// 3. Update cbuild_workspace.yml
	ws := &ccommon.Workspace{}
	err = ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace for update: %w", err)
	}

	if ws.Targets == nil {
		ws.Targets = make(map[string]*ccommon.TargetConfiguration)
	}

	if _, exists := ws.Targets[destName]; exists {
		fmt.Printf("TargetConfiguration %s already exists in cbuild_workspace.yml. Skipping update.\n", destName)
	} else {
		ws.Targets[destName] = &ccommon.TargetConfiguration{
			ProjectType: "CMake",
		}
		err = ws.Save()
		if err != nil {
			return fmt.Errorf("error saving updated workspace: %w", err)
		}
		fmt.Printf("Added target %s to cbuild_workspace.yml.\n", destName)
	}

	// 4. Process csetup.yml if it exists
	err = ws.ProcessCSetupFile(destName)
	if err != nil {
		return fmt.Errorf("error processing csetup file: %w", err)
	}

	fmt.Println("Repository cloned successfully.")
	return nil
}
