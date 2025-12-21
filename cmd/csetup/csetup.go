package main

import (
	"flag"
	"fmt"
	"os"
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
