package main

import (
	"cbuild-go/pkg/ccommon"
	"cbuild-go/pkg/cli"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	flags := []cli.Flag{
		ccommon.WorkspaceFlag,
		ccommon.ConfigFlag,
		ccommon.ToolchainFlag,
		ccommon.ReinitFlag,
		ccommon.DownloadDepsFlag,
		ccommon.DeleteFlag,
	}

	ctx, args, parseErr := cli.ParseFlags(context.Background(), os.Args[1:], flags)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", parseErr)
		os.Exit(1)
	}

	workspacePath := cli.GetString(ctx, cli.FlagKey(ccommon.FlagWorkspace))
	if workspacePath == "" {
		workspacePath = "."
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

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
		err = handleInit(ctx, workspacePath, subArgs)
	case "git-clone":
		err = handleGitClone(ctx, workspacePath, subArgs)
	case "add-dependency":
		err = handleAddDependency(ctx, workspacePath, subArgs)
	case "remove-dependency":
		err = handleRemoveDependency(ctx, workspacePath, subArgs)
	case "remove-source":
		err = handleRemoveSource(ctx, workspacePath, subArgs)
	case "set-cxx-version":
		err = handleSetCXXVersion(ctx, workspacePath, subArgs)
	case "enable-staging":
		err = handleEnableStaging(ctx, workspacePath, subArgs)
	case "disable-staging":
		err = handleDisableStaging(ctx, workspacePath, subArgs)
	case "list-sources":
		err = handleListSources(ctx, workspacePath, subArgs)
	case "get-args":
		err = handleGetArgs(ctx, workspacePath, subArgs)
	case "detect-toolchains":
		err = handleDetectToolchains(ctx, workspacePath, subArgs)
	case "add-config":
		err = handleAddConfig(ctx, workspacePath, subArgs)
	case "remove-config":
		err = handleRemoveConfig(ctx, workspacePath, subArgs)
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
