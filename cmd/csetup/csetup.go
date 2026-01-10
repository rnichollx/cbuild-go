package main

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
	"os"
	"path/filepath"
)

func getWorkspacePath(ctx context.Context) string {
	workspacePath := cli.GetString(ctx, cli.FlagKey(ccommon.FlagWorkspace))
	if workspacePath == "" {
		workspacePath = "."
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
	return workspacePath
}

func main() {
	runner := &cli.Runner{
		Name:        "csetup",
		Description: "Workspace setup tool for cbuild",
		GlobalFlags: []cli.Flag{
			ccommon.WorkspaceFlag,
			ccommon.HelpFlag,
		},
		Subcommands: make(map[string]*cli.Subcommand),
	}

	runner.Subcommands["init"] = &cli.Subcommand{
		Description:  "Initialize a new workspace",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ReinitFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleInit(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["git-clone"] = &cli.Subcommand{
		Description:  "Clone a git repository into the workspace",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.DownloadDepsFlag, ccommon.SubmoduleFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGitClone(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["add-dependency"] = &cli.Subcommand{
		Description: "Add a dependency to a source",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleAddDependency(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["remove-dependency"] = &cli.Subcommand{
		Description: "Remove a dependency from a source",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveDependency(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["remove-source"] = &cli.Subcommand{
		Description:  "Remove a source from the workspace",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.DeleteFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveSource(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["remove-target"] = &cli.Subcommand{
		Description: "Remove a target from the workspace",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveTarget(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["remove-project"] = &cli.Subcommand{
		Description:  "Remove a source and all its associated targets from the workspace",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.DeleteFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveProject(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["set-cxx-version"] = &cli.Subcommand{
		Description: "Set the C++ version for a source or the whole workspace",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleSetCXXVersion(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["enable-staging"] = &cli.Subcommand{
		Description: "Enable staging for a source",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleEnableStaging(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["disable-staging"] = &cli.Subcommand{
		Description: "Disable staging for a source",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleDisableStaging(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["list-sources"] = &cli.Subcommand{
		Description: "List all sources in the workspace",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleListSources(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["get-args"] = &cli.Subcommand{
		Description:  "Get build arguments for a target",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGetArgs(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["detect-toolchains"] = &cli.Subcommand{
		Description: "Detect system toolchains",
		AllowArgs:   true,
		Exec: func(ctx context.Context, args []string) error {
			return handleDetectToolchains(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["add-config"] = &cli.Subcommand{
		Description:  "Add a build configuration",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddConfig(ctx, getWorkspacePath(ctx), args)
		},
	}
	runner.Subcommands["remove-config"] = &cli.Subcommand{
		Description:  "Remove a build configuration",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveConfig(ctx, getWorkspacePath(ctx), args)
		},
	}

	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
