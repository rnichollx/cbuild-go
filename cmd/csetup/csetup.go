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
					absPath, err := filepath.Abs(curr)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: failed to get absolute path: %v\n", err)
						os.Exit(1)
					}
					workspacePath = absPath
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

	var err error
	switch subcommand {
	case "init":
		err = handleInit(workspacePath, subArgs)
	case "git-clone":
		err = handleGitClone(workspacePath, subArgs)
	case "add-dependency":
		err = handleAddDependency(workspacePath, subArgs)
	case "remove-dependency":
		err = handleRemoveDependency(workspacePath, subArgs)
	case "remove-source":
		err = handleRemoveSource(workspacePath, subArgs)
	case "set-cxx-version":
		err = handleSetCXXVersion(workspacePath, subArgs)
	case "enable-staging":
		err = handleEnableStaging(workspacePath, subArgs)
	case "disable-staging":
		err = handleDisableStaging(workspacePath, subArgs)
	case "list-sources":
		err = handleListSources(workspacePath, subArgs)
	case "get-args":
		err = handleGetArgs(workspacePath, subArgs)
	case "detect-toolchains":
		err = handleDetectToolchains(workspacePath, subArgs)
	case "add-config":
		err = handleAddConfig(workspacePath, subArgs)
	case "remove-config":
		err = handleRemoveConfig(workspacePath, subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: csetup [-w workspace] <subcommand> [args]")
	fmt.Println("Subcommands:")
	fmt.Println("  init <workspace name> [--reinit]")
	fmt.Println("  git-clone <repo_url> [dest_name] [--download-deps]")
	fmt.Println("  add-dependency <source> <depname>")
	fmt.Println("  remove-dependency <source> <depname>")
	fmt.Println("  remove-source <source> [-D|--delete]")
	fmt.Println("  set-cxx-version <version> [<source>]")
	fmt.Println("  enable-staging <source>")
	fmt.Println("  disable-staging <source>")
	fmt.Println("  list-sources")
	fmt.Println("  get-args <target> [-T|--toolchain <toolchain>] [-c|--config <type>]")
	fmt.Println("  detect-toolchains")
	fmt.Println("  add-config <configname>")
	fmt.Println("  remove-config <configname>")
}
