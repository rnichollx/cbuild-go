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
	if _, err := os.Stat(workspaceConfig); err == nil && !reinit {
		fmt.Fprintf(os.Stderr, "Error: %s already exists. Use --reinit to overwrite.\n", workspaceConfig)
		os.Exit(1)
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
