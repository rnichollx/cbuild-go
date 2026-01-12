package csetupapp

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

var CSetup = &cli.Runner{
	Name:        "csetup",
	Description: "Workspace setup tool for cbuild",
	GlobalFlags: []cli.Flag{
		ccommon.WorkspaceFlag,
		ccommon.HelpFlag,
	},
	Subcommands: make(map[string]*cli.Subcommand),
}

func init() {
	CSetup.Subcommands["init"] = &cli.Subcommand{
		Description: "Initialize a new workspace",
		Arguments: []cli.Argument{
			{Name: "path", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.ReinitFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleInit(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["git-clone"] = &cli.Subcommand{
		Description: "Clone a git repository into the workspace",
		Arguments: []cli.Argument{
			{Name: "url", Required: true},
			{Name: "path", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.DownloadDepsFlag, ccommon.SubmoduleFlag, ccommon.NoSetupFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGitClone(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["download"] = &cli.Subcommand{
		Description: "Download missing sources",
		Arguments: []cli.Argument{
			{Name: "source", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.DownloadDepsFlag, ccommon.NoSetupFlag, ccommon.SubmoduleFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDownload(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["load-defaults"] = &cli.Subcommand{
		Description: "Load default configuration for a source from its csetup.yml",
		Arguments: []cli.Argument{
			{Name: "source", Required: true},
		},
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleLoadDefaults(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["add-dependency"] = &cli.Subcommand{
		Description: "Add a dependency to a source",
		Arguments: []cli.Argument{
			{Name: "source", Required: true},
			{Name: "dependency", Required: true},
		},
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleAddDependency(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["remove-dependency"] = &cli.Subcommand{
		Description: "Remove a dependency from a source",
		Arguments: []cli.Argument{
			{Name: "source", Required: true},
			{Name: "dependency", Required: true},
		},
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveDependency(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["remove-source"] = &cli.Subcommand{
		Description: "Remove a source from the workspace",
		Arguments: []cli.Argument{
			{Name: "source", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.SourceFlag, ccommon.DeleteFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveSource(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["remove-target"] = &cli.Subcommand{
		Description: "Remove a target from the workspace",
		Arguments: []cli.Argument{
			{Name: "target", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveTarget(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["remove-project"] = &cli.Subcommand{
		Description: "Remove a source and all its associated targets from the workspace",
		Arguments: []cli.Argument{
			{Name: "source", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.SourceFlag, ccommon.DeleteFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveProject(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["set-cxx-version"] = &cli.Subcommand{
		Description: "Set the C++ version for a source or the whole workspace",
		Arguments: []cli.Argument{
			{Name: "source", Required: false},
			{Name: "version", Required: true},
		},
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleSetCXXVersion(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["enable-staging"] = &cli.Subcommand{
		Description: "Enable staging for a source",
		Arguments: []cli.Argument{
			{Name: "source", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.SourceFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleEnableStaging(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["disable-staging"] = &cli.Subcommand{
		Description: "Disable staging for a source",
		Arguments: []cli.Argument{
			{Name: "source", Required: false},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.SourceFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDisableStaging(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["list-sources"] = &cli.Subcommand{
		Description:           "List all sources in the workspace",
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleListSources(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["drop-files"] = &cli.Subcommand{
		Description: "Delete local source files without removing them from configuration",
		Arguments: []cli.Argument{
			{Name: "source", Required: true},
		},
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleDropFiles(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["get-args"] = &cli.Subcommand{
		Description: "Get build arguments for a target",
		Arguments: []cli.Argument{
			{Name: "target", Required: true},
		},
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGetArgs(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["detect-toolchains"] = &cli.Subcommand{
		Description:           "Detect system toolchains",
		AllowUnrecognizedArgs: true,
		Exec: func(ctx context.Context, args []string) error {
			return handleDetectToolchains(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["add-config"] = &cli.Subcommand{
		Description:           "Add a build configuration",
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddConfig(ctx, getWorkspacePath(ctx), args)
		},
	}
	CSetup.Subcommands["remove-config"] = &cli.Subcommand{
		Description:           "Remove a build configuration",
		AllowUnrecognizedArgs: true,
		AcceptsFlags:          []cli.Flag{ccommon.ConfigFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveConfig(ctx, getWorkspacePath(ctx), args)
		},
	}
}
