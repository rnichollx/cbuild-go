package main

import (
	"cbuild-go/pkg/ccommon"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var workspacePath string

	// Common flags
	fs := flag.NewFlagSet("csetup", flag.ExitOnError)
	fs.StringVar(&workspacePath, "w", ".", "path to the workspace directory")
	fs.StringVar(&workspacePath, "workspace", ".", "path to the workspace directory")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// We need to parse common flags that might appear before or after the subcommand
	// But the requirement says "csetup git-clone ...", so let's check for subcommand first.

	subcommand := ""
	subArgs := []string{}

	// Basic parsing to handle flags before/after subcommand
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if arg == "-w" || arg == "--workspace" {
				if i+1 < len(args) {
					workspacePath = args[i+1]
					i++
				} else {
					fmt.Fprintf(os.Stderr, "Error: flag %s requires an argument\n", arg)
					os.Exit(1)
				}
			} else {
				// Ignore other flags for now or handle them
			}
		} else if subcommand == "" {
			subcommand = arg
			subArgs = args[i+1:]
			break
		}
	}

	if subcommand == "" {
		printUsage()
		os.Exit(1)
	}

	// Try to find the workspace by looking up from current directory
	if workspacePath == "." {
		cwd, err := os.Getwd()
		if err == nil {
			curr := cwd
			for {
				if _, err := os.Stat(filepath.Join(curr, "cbuild_workspace.yml")); err == nil {
					workspacePath, _ = filepath.Abs(curr)
					break
				}
				parent := filepath.Dir(curr)
				if parent == curr {
					break
				}
				curr = parent
			}
		}
	} else {
		absPath, err := filepath.Abs(workspacePath)
		if err == nil {
			workspacePath = absPath
		}
	}

	switch subcommand {
	case "init":
		handleInit(workspacePath, subArgs)
	case "git-clone":
		handleGitClone(workspacePath, subArgs)
	case "add-dependency":
		handleAddDependency(workspacePath, subArgs)
	case "remove-dependency":
		handleRemoveDependency(workspacePath, subArgs)
	case "remove-source":
		handleRemoveSource(workspacePath, subArgs)
	case "set-cxx-version":
		handleSetCXXVersion(workspacePath, subArgs)
	case "enable-staging":
		handleEnableStaging(workspacePath, subArgs)
	case "disable-staging":
		handleDisableStaging(workspacePath, subArgs)
	case "list-sources":
		handleListSources(workspacePath, subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: csetup [-w workspace] <subcommand> [args]")
	fmt.Println("Subcommands:")
	fmt.Println("  init <workspace name> [--reinit]")
	fmt.Println("  git-clone <repo_url> [dest_name]")
	fmt.Println("  add-dependency <source> <depname>")
	fmt.Println("  remove-dependency <source> <depname>")
	fmt.Println("  remove-source <source> [-D|--delete]")
	fmt.Println("  set-cxx-version <source> <version>")
	fmt.Println("  enable-staging <source>")
	fmt.Println("  disable-staging <source>")
	fmt.Println("  list-sources")
}

func handleListSources(workspacePath string, args []string) {
	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	sourcesDir := filepath.Join(workspacePath, "sources")
	dirEntries, err := os.ReadDir(sourcesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading sources directory: %v\n", err)
		os.Exit(1)
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
}

func handleAddDependency(workspacePath string, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: csetup add-dependency <source> <depname>")
		os.Exit(1)
	}

	source := args[0]
	depname := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	target, ok := ws.Targets[source]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
	}

	// Check if dependency already exists
	for _, d := range target.Depends {
		if d == depname {
			fmt.Printf("Dependency %s already exists for %s\n", depname, source)
			return
		}
	}

	target.Depends = append(target.Depends, depname)
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added dependency %s to %s\n", depname, source)
}

func handleRemoveDependency(workspacePath string, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: csetup remove-dependency <source> <depname>")
		os.Exit(1)
	}

	source := args[0]
	depname := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	target, ok := ws.Targets[source]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
	}

	newDepends := []string{}
	found := false
	for _, d := range target.Depends {
		if d == depname {
			found = true
			continue
		}
		newDepends = append(newDepends, d)
	}

	if !found {
		fmt.Printf("Dependency %s not found for %s\n", depname, source)
		return
	}

	target.Depends = newDepends
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed dependency %s from %s\n", depname, source)
}

func handleRemoveSource(workspacePath string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: csetup remove-source <source> [-D|--delete]")
		os.Exit(1)
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
		fmt.Println("Usage: csetup remove-source <source> [-D|--delete]")
		os.Exit(1)
	}

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	if _, ok := ws.Targets[source]; !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
	}

	delete(ws.Targets, source)
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed source %s from workspace\n", source)

	if removeFolder {
		sourceDir := filepath.Join(workspacePath, "sources", source)
		if _, err := os.Stat(sourceDir); err == nil {
			fmt.Printf("Deleting source folder: %s\n", sourceDir)
			err = os.RemoveAll(sourceDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting source folder: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Source folder %s not found, skipping deletion.\n", sourceDir)
		}
	} else {
		fmt.Printf("Note: files in sources/%s were NOT deleted. Use -R to delete them.\n", source)
	}
}

func handleSetCXXVersion(workspacePath string, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: csetup set-cxx-version <source> <version>")
		os.Exit(1)
	}

	// Wait, the instruction says set-cxx-version <source> <version>
	// But Workspace has CXXVersion, not Target.
	// Does the user mean set workspace CXXVersion or Target CXXVersion?
	// Workspace has CXXVersion (line 47 of pkg/ccommon/ccommon.go)
	// Target does NOT have CXXVersion.

	// If the command is <source> <version>, maybe it's meant to be global if source is "global"?
	// Or maybe I should add CXXVersion to Target?

	// Looking at pkg/ccommon/ccommon.go:
	// type Workspace struct {
	//    ...
	//    CXXVersion  string  `yaml:"cxx_version"`
	// }

	// I'll assume for now it's a global setting if they provide a source that matches,
	// but the command format suggests it's per source.
	// Let's check if I should add CXXVersion to Target.

	// Actually, maybe <source> is ignored or it's meant to be workspace-wide and <source> is just a placeholder or I should check if it's "workspace".

	source := args[0]
	version := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	// For now, let's just set it on the workspace regardless of 'source'
	// OR maybe 'source' is intended to be the workspace name?
	// The other commands use 'source' to refer to a target.

	ws.CXXVersion = version
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Set CXX version to %s (source argument %s was ignored as it is currently global)\n", version, source)
}

func handleEnableStaging(workspacePath string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: csetup enable-staging <source>")
		os.Exit(1)
	}

	source := args[0]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	target, ok := ws.Targets[source]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
	}

	staged := true
	target.Staged = &staged

	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Enabled staging for %s\n", source)
}

func handleDisableStaging(workspacePath string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: csetup disable-staging <source>")
		os.Exit(1)
	}

	source := args[0]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	target, ok := ws.Targets[source]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
	}

	staged := false
	target.Staged = &staged

	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Disabled staging for %s\n", source)
}

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

func handleGitClone(workspacePath string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: csetup git-clone <repo_url> [dest_name]")
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error: %s not found. csetup must be run in a cbuild workspace.\n", workspaceConfig)
		os.Exit(1)
	}

	// 2. Clone into sources/<destName>
	destDir := filepath.Join(workspacePath, "sources", destName)

	fmt.Printf("Cloning %s into %s...\n", repoURL, destDir)

	cmd := exec.Command("git", "clone", repoURL, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	// 3. Update cbuild_workspace.yml
	ws := &ccommon.Workspace{}
	err = ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace for update: %v\n", err)
		os.Exit(1)
	}

	if ws.Targets == nil {
		ws.Targets = make(map[string]*ccommon.Target)
	}

	if _, exists := ws.Targets[destName]; exists {
		fmt.Printf("Target %s already exists in cbuild_workspace.yml. Skipping update.\n", destName)
	} else {
		ws.Targets[destName] = &ccommon.Target{
			ProjectType: "CMake",
		}
		err = ws.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving updated workspace: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added target %s to cbuild_workspace.yml.\n", destName)
	}

	// 4. Process csetup.yml if it exists
	err = ws.ProcessCSetupFile(destName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing csetup file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository cloned successfully.")
}
