package csetupapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func getWorkspacePath(ctx context.Context) string {
	workspacePathRaw, _ := cli.GetPath(ctx, ccommon.PWorkspace)
	workspacePath := ""
	if workspacePathRaw != nil {
		workspacePath = *workspacePathRaw
	}
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

var (
	PPath       = cli.NewParameter("path", cli.ParameterTypePath, nil, "path to the workspace or source", false)
	PUrl        = cli.NewParameter("url", cli.ParameterTypeURI, nil, "URL of the repository", true)
	PLocalPath  = cli.NewParameter("local_path", cli.ParameterTypePath, nil, "path to the local source", true)
	PSource     = cli.NewParameter("source", cli.ParameterTypeString, nil, "name of the source", false)
	PSourceReq  = cli.NewParameter("source", cli.ParameterTypeString, nil, "name of the source", true)
	PDependency = cli.NewParameter("dependency", cli.ParameterTypeString, nil, "name of the dependency", true)
	PTarget     = cli.NewParameter("target", cli.ParameterTypeString, nil, "Name of the build target.", false)
	PTargetReq  = cli.NewParameter("target", cli.ParameterTypeString, nil, "name of the target", true)
)

func init() {
	CSetup.Subcommands["init"] = &cli.Subcommand{
		Description: "Initialize a new workspace",
		Arguments: []cli.Argument{
			cli.NewStringArgument("path", PPath),
		},
		AcceptsFlags: []cli.Flag{ccommon.ReinitFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleInit(ctx)
		},
	}
	CSetup.Subcommands["git-clone"] = &cli.Subcommand{
		Description: "Clone a git repository into the workspace",
		Arguments: []cli.Argument{
			cli.NewStringArgument("url", PUrl),
			cli.NewStringArgument("path", PPath),
		},
		AcceptsFlags: []cli.Flag{ccommon.DownloadDepsFlag, ccommon.SubmoduleFlag, ccommon.NoSetupFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGitClone(ctx)
		},
	}
	CSetup.Subcommands["declare-git-source"] = &cli.Subcommand{
		Description: "Add git source information to the workspace without downloading",
		Arguments: []cli.Argument{
			cli.NewStringArgument("url", PUrl),
			cli.NewStringArgument("path", PPath),
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleDeclareGitSource(ctx)
		},
	}
	CSetup.Subcommands["declare-local-source"] = &cli.Subcommand{
		Description: "Add local source information to the workspace",
		Arguments: []cli.Argument{
			cli.NewStringArgument("local_path", PLocalPath),
			cli.NewStringArgument("path", PPath),
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleDeclareLocalSource(ctx)
		},
	}
	CSetup.Subcommands["download"] = &cli.Subcommand{
		Description: "Download missing sources",
		Arguments: []cli.Argument{
			cli.NewStringArgument("source", PSource),
		},
		AcceptsFlags: []cli.Flag{ccommon.DownloadDepsFlag, ccommon.NoSetupFlag, ccommon.SubmoduleFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDownload(ctx)
		},
	}
	CSetup.Subcommands["load-defaults"] = &cli.Subcommand{
		Description: "Load default configuration for a source from its csetup.yml",
		Arguments: []cli.Argument{
			cli.NewStringArgument("source", PSourceReq),
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleLoadDefaults(ctx)
		},
	}
	CSetup.Subcommands["add-dependency"] = &cli.Subcommand{
		Description: "Add a dependency to a target",
		Arguments: []cli.Argument{
			cli.NewStringArgument("target", PTargetReq),
			cli.NewStringArgument("dependency", PDependency),
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddDependency(ctx)
		},
	}
	CSetup.Subcommands["remove-dependency"] = &cli.Subcommand{
		Description: "Remove a dependency from a target",
		Arguments: []cli.Argument{
			cli.NewStringArgument("target", PTargetReq),
			cli.NewStringArgument("dependency", PDependency),
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveDependency(ctx)
		},
	}
	CSetup.Subcommands["remove-source"] = &cli.Subcommand{
		Description: "Remove a source from the workspace",
		Arguments: []cli.Argument{
			ccommon.SourceArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.DeleteFlag, ccommon.SourceFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveSource(ctx)
		},
	}
	CSetup.Subcommands["remove-target"] = &cli.Subcommand{
		Description: "Remove a target from the workspace",
		Arguments: []cli.Argument{
			ccommon.TargetArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveTarget(ctx)
		},
	}
	CSetup.Subcommands["remove-project"] = &cli.Subcommand{
		Description: "Remove a source and all its associated targets from the workspace",
		Arguments: []cli.Argument{
			ccommon.SourceArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.DeleteFlag, ccommon.SourceFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveProject(ctx)
		},
	}
	CSetup.Subcommands["tidy"] = &cli.Subcommand{
		Description:  "Delete source folders that are not in the sources list",
		AcceptsFlags: []cli.Flag{ccommon.DryRunFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleTidy(ctx)
		},
	}
	CSetup.Subcommands["new-target"] = &cli.Subcommand{
		Description: "Add a new target to the workspace",
		Arguments: []cli.Argument{
			ccommon.TargetArg,
			ccommon.SourceArg,
		},
		AcceptsFlags: []cli.Flag{
			ccommon.SourceFlag,
			ccommon.TargetFlag,
			ccommon.OverwriteFlag,
			ccommon.ProjectTypeFlag,
			ccommon.CMakePackageNameFlag,
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleNewTarget(ctx)
		},
	}
	CSetup.Subcommands["set-cxx-version"] = &cli.Subcommand{
		Description: "Set the C++ version for a target or the whole workspace",
		Arguments: []cli.Argument{
			ccommon.CxxVersionArg,
			ccommon.TargetArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleSetCXXVersion(ctx)
		},
	}
	CSetup.Subcommands["enable-staging"] = &cli.Subcommand{
		Description: "Enable staging for a target",
		Arguments: []cli.Argument{
			cli.NewStringArgument("target", PTarget),
			ccommon.TargetArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleEnableStaging(ctx)
		},
	}
	CSetup.Subcommands["disable-staging"] = &cli.Subcommand{
		Description: "Disable staging for a target",
		Arguments: []cli.Argument{
			cli.NewStringArgument("target", PTarget),
			ccommon.TargetArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDisableStaging(ctx)
		},
	}
	CSetup.Subcommands["list-sources"] = &cli.Subcommand{
		Description: "List all sources in the workspace",
		Exec: func(ctx context.Context, args []string) error {
			return handleListSources(ctx)
		},
	}
	CSetup.Subcommands["drop-files"] = &cli.Subcommand{
		Description: "Delete local source files without removing them from configuration",
		Arguments: []cli.Argument{
			cli.NewStringArgument("source", PSourceReq),
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleDropFiles(ctx)
		},
	}
	CSetup.Subcommands["get-args"] = &cli.Subcommand{
		Description: "Get build arguments for a target",
		Arguments: []cli.Argument{
			cli.NewStringArgument("target", PTargetReq),
			ccommon.TargetArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGetArgs(ctx)
		},
	}
	CSetup.Subcommands["detect-toolchains"] = &cli.Subcommand{
		Description:  "Detect system toolchains",
		AcceptsFlags: []cli.Flag{ccommon.DebugFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDetectToolchains(ctx)
		},
	}
	CSetup.Subcommands["add-config"] = &cli.Subcommand{
		Description:  "Add a build configuration",
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddConfig(ctx)
		},
	}
	CSetup.Subcommands["remove-config"] = &cli.Subcommand{
		Description:  "Remove a build configuration",
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveConfig(ctx)
		},
	}
}
