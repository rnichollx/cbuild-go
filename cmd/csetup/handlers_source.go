package main

import (
	"context"
	"fmt"
	"os"
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
		return fmt.Errorf("usage: csetup git-clone <repo_url> [dest_name] [--download-deps] [--submodule] [--no-setup]")
	}

	repoURL := ""
	destName := ""
	downloadDeps := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDownload))
	noSetup := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagNoSetup))

	for _, arg := range args {
		if repoURL == "" {
			repoURL = arg
		} else if destName == "" {
			destName = arg
		}
	}

	if repoURL == "" {
		return fmt.Errorf("usage: csetup git-clone <repo_url> [dest_name] [--download-deps] [--submodule] [--no-setup]")
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

	ws := &ccommon.WorkspaceContext{}
	ws.DownloadDeps = downloadDeps
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	// 2. Clone into sources/<destName>
	if ws.Config.Sources == nil {
		ws.Config.Sources = make(map[string]*ccommon.CodeSource)
	}
	ws.Config.Sources[destName] = &ccommon.CodeSource{
		Git: &ccommon.GitSource{
			Repository: repoURL,
		},
	}

	err = ws.DownloadSource(ctx, destName)
	if err != nil {
		return fmt.Errorf("error downloading source: %w", err)
	}

	// 3. Update cbuild_workspace.yml with target if it doesn't exist
	targetExists := false
	if ws.Config.Targets == nil {
		ws.Config.Targets = make(map[string]*ccommon.TargetConfiguration)
	}

	if _, ok := ws.Config.Targets[destName]; ok {
		fmt.Printf("Target %s already exists in cbuild_workspace.yml. Skipping target creation.\n", destName)
		targetExists = true
	} else {
		ws.Config.Targets[destName] = &ccommon.TargetConfiguration{
			Source: destName,
		}
		err = ws.Save()
		if err != nil {
			return fmt.Errorf("error saving updated workspace: %w", err)
		}
		fmt.Printf("Added target %s to cbuild_workspace.yml.\n", destName)
	}

	// 4. Process csetup.yml if it exists and setup is not disabled
	if !noSetup && !targetExists {
		err = ws.ProcessCSetupConfig(ctx, destName)
		if err != nil {
			return fmt.Errorf("error processing csetup file: %w", err)
		}
	} else if noSetup {
		fmt.Println("Skipping setup as requested by --no-setup.")
	} else if targetExists {
		fmt.Printf("Skipping setup as target %s already exists.\n", destName)
	}

	fmt.Println("Repository cloned successfully.")
	return nil
}

func handleDownload(ctx context.Context, workspacePath string, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: csetup download [source_name] [--download-deps] [--no-setup] [--submodule]")
	}

	downloadDeps := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDownload))
	noSetup := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagNoSetup))

	ws := &ccommon.WorkspaceContext{}
	ws.DownloadDeps = downloadDeps
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	sourcesToDownload := []string{}
	if len(args) == 1 {
		sourceName := args[0]
		if _, ok := ws.Config.Sources[sourceName]; !ok {
			return fmt.Errorf("source %s not found in workspace configuration", sourceName)
		}
		sourcesToDownload = append(sourcesToDownload, sourceName)
	} else {
		for name := range ws.Config.Sources {
			sourcesToDownload = append(sourcesToDownload, name)
		}
	}

	for _, sourceName := range sourcesToDownload {
		sourceDir := filepath.Join(workspacePath, "sources", sourceName)
		if _, err := os.Stat(sourceDir); err == nil {
			if len(args) == 1 {
				fmt.Printf("Source %s already exists at %s\n", sourceName, sourceDir)
			}
			continue
		}

		err = ws.DownloadSource(ctx, sourceName)
		if err != nil {
			return fmt.Errorf("error downloading source %s: %w", sourceName, err)
		}

		if !noSetup {
			err = ws.ProcessCSetupConfig(ctx, sourceName)
			if err != nil {
				fmt.Printf("Warning: error processing csetup for %s: %v\n", sourceName, err)
			}
		}
	}

	return nil
}

func handleLoadDefaults(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: csetup load-defaults <source>")
	}

	sourceName := args[0]
	if sourceName == "" {
		return fmt.Errorf("usage: csetup load-defaults <source>")
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

	// Create a new target with that source name if it doesn't exist
	if ws.Config.Targets == nil {
		ws.Config.Targets = make(map[string]*ccommon.TargetConfiguration)
	}

	if _, ok := ws.Config.Targets[sourceName]; !ok {
		ws.Config.Targets[sourceName] = &ccommon.TargetConfiguration{
			Source: sourceName,
		}
		fmt.Printf("Created target %s for source %s\n", sourceName, sourceName)
	}

	err = ws.ProcessCSetupConfig(ctx, sourceName)
	if err != nil {
		return fmt.Errorf("error processing csetup file: %w", err)
	}

	fmt.Printf("Defaults loaded for source %s\n", sourceName)
	return nil
}
