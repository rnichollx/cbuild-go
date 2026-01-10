package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleListSources(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("Unexpected args: %v", args)
	}
	ws := &ccommon.WorkspaceContext{}
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
	for name := range ws.Config.Targets {
		status := "[MISSING]"
		target, err := ws.GetTarget(name)
		if err != nil {
			return err
		}
		srcPath, err := target.CMakeSourcePath(ws)
		if err == nil {
			if info, err := os.Stat(srcPath); err == nil && info.IsDir() {
				if target.Config.ExternalSourceOverride != nil {
					status = "[OK EXTERNAL]"
				} else {
					status = "[OK]"
				}
			}
		}

		fmt.Printf("%s %s\n", name, status)
		// If it's a standard source (in the sources/ directory), mark it as tracked
		if target.Config.ExternalSourceOverride == nil {
			sourceName := target.Config.Source
			if sourceName == "" {
				sourceName = name
			}
			trackedDirs[sourceName] = true
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

func handleRemoveSource(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: csetup remove-source <source> [-X|--delete]")
	}

	sourceName := args[0]
	removeFolder := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDelete))

	if sourceName == "" {
		return fmt.Errorf("usage: csetup remove-source <source> [-X|--delete]")
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	if ws.Config.Sources == nil {
		return fmt.Errorf("no sources defined in workspace")
	}

	if _, ok := ws.Config.Sources[sourceName]; !ok {
		return fmt.Errorf("source %s not found in workspace", sourceName)
	}

	delete(ws.Config.Sources, sourceName)

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed source %s from workspace\n", sourceName)

	if removeFolder {
		sourceDir := filepath.Join(workspacePath, "sources", sourceName)
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
		fmt.Printf("Note: files in sources/%s were NOT deleted. Use -X to delete them.\n", sourceName)
	}
	return nil
}

func handleRemoveTarget(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: csetup remove-target <target>")
	}

	targetName := args[0]
	if targetName == "" {
		return fmt.Errorf("usage: csetup remove-target <target>")
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	if _, ok := ws.Config.Targets[targetName]; !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	delete(ws.Config.Targets, targetName)

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed target %s from workspace\n", targetName)
	return nil
}

func handleRemoveProject(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: csetup remove-project <source> [-X|--delete]")
	}

	sourceName := args[0]
	removeFolder := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDelete))

	if sourceName == "" {
		return fmt.Errorf("usage: csetup remove-project <source> [-X|--delete]")
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	sourceFound := false
	if ws.Config.Sources != nil {
		if _, ok := ws.Config.Sources[sourceName]; ok {
			delete(ws.Config.Sources, sourceName)
			sourceFound = true
		}
	}

	targetsToRemove := []string{}
	for targetName, targetConfig := range ws.Config.Targets {
		targetSource := targetConfig.Source
		if targetSource == "" {
			targetSource = targetName
		}
		if targetSource == sourceName {
			targetsToRemove = append(targetsToRemove, targetName)
		}
	}

	for _, targetName := range targetsToRemove {
		delete(ws.Config.Targets, targetName)
	}

	if !sourceFound && len(targetsToRemove) == 0 {
		return fmt.Errorf("source or targets for %s not found in workspace", sourceName)
	}

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	if sourceFound {
		fmt.Printf("Removed source %s from workspace\n", sourceName)
	}
	for _, targetName := range targetsToRemove {
		fmt.Printf("Removed target %s from workspace\n", targetName)
	}

	if removeFolder {
		sourceDir := filepath.Join(workspacePath, "sources", sourceName)
		if _, err := os.Stat(sourceDir); err == nil {
			fmt.Printf("Deleting source folder: %s\n", sourceDir)
			err = os.RemoveAll(sourceDir)
			if err != nil {
				return fmt.Errorf("error deleting source folder: %w", err)
			}
		} else {
			fmt.Printf("Source folder %s not found, skipping deletion.\n", sourceDir)
		}
	} else if sourceFound {
		fmt.Printf("Note: files in sources/%s were NOT deleted. Use -X to delete them.\n", sourceName)
	}

	return nil
}

func handleGitClone(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: csetup git-clone <repo_url> [dest_name] [--download-deps] [--submodule]")
	}

	repoURL := ""
	destName := ""
	downloadDeps := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDownload))
	useSubmodule := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagSubmodule))

	for _, arg := range args {
		if repoURL == "" {
			repoURL = arg
		} else if destName == "" {
			destName = arg
		}
	}

	if repoURL == "" {
		return fmt.Errorf("usage: csetup git-clone <repo_url> [dest_name] [--download-deps] [--submodule]")
	}

	if destName == "" {
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

	if useSubmodule {
		fmt.Printf("Adding submodule %s into %s...\n", repoURL, destDir)
	} else {
		fmt.Printf("Cloning %s into %s...\n", repoURL, destDir)
	}

	var cmd *exec.Cmd
	if useSubmodule {
		// Use relative path for submodule add to ensure .gitmodules is correct
		relDestDir := filepath.Join("sources", destName)
		cmd = exec.Command("git", "submodule", "add", repoURL, relDestDir)
	} else {
		cmd = exec.Command("git", "clone", repoURL, destDir)
	}
	cmd.Dir = workspacePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if useSubmodule {
			return fmt.Errorf("error adding submodule: %w", err)
		}
		return fmt.Errorf("error cloning repository: %w", err)
	}

	// 3. Update cbuild_workspace.yml
	ws := &ccommon.WorkspaceContext{}
	ws.DownloadDeps = downloadDeps
	err = ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace for update: %w", err)
	}

	if ws.Config.Targets == nil {
		ws.Config.Targets = make(map[string]*ccommon.TargetConfiguration)
	}

	if _, ok := ws.Config.Targets[destName]; ok {
		fmt.Printf("TargetConfiguration %s already exists in cbuild_workspace.yml. Skipping update.\n", destName)
	} else {
		if ws.Config.Sources == nil {
			ws.Config.Sources = make(map[string]*ccommon.CodeSource)
		}
		ws.Config.Sources[destName] = &ccommon.CodeSource{
			Git: &ccommon.GitSource{
				Repository: repoURL,
			},
		}

		ws.Config.Targets[destName] = &ccommon.TargetConfiguration{
			ProjectType: "CMake",
			Source:      destName,
		}
		err = ws.Save()
		if err != nil {
			return fmt.Errorf("error saving updated workspace: %w", err)
		}
		fmt.Printf("Added target %s to cbuild_workspace.yml.\n", destName)
	}

	// 4. Process csetup.yml if it exists
	err = ws.ProcessCSetupFile(ctx, destName)
	if err != nil {
		return fmt.Errorf("error processing csetup file: %w", err)
	}

	fmt.Println("Repository cloned successfully.")
	return nil
}
