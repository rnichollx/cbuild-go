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
	workspacePathRaw := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter)
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
	URLParameter            = cli.Parameter{Key: "url", Type: cli.ParameterTypeURI, DefaultValue: nil, Description: "URL of the repository"}
	SourcePathParameter     = cli.Parameter{Key: "sourcepath", Type: cli.ParameterTypePath, DefaultValue: nil, Description: "path to the local source"}
	SourcePathFlag          = cli.Flag{Short: "", Long: "sourcepath", Parameter: SourcePathParameter}
	HelpSubcommandParameter = cli.Parameter{Key: "help-subcommand", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "subcommand to show help for"}
	SourceParameter         = cli.Parameter{Key: "source", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "name of the source"}
	DependencyParameter     = cli.Parameter{Key: "dependency", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "name of the dependency"}
	TargetParameter         = cli.Parameter{Key: "target", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "Name of the build target."}
	RevisionParameter       = cli.Parameter{Key: "revision", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "Revision or tag to use."}
	BranchParameter         = cli.Parameter{Key: "branch", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "Branch name to use."}
	NoTrackParameter        = cli.Parameter{Key: "no-track", Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Do not track the cloned branch."}
)

func init() {
	CSetup.Subcommands["help"] = &cli.Subcommand{
		Description: "Show help for csetup or a specific subcommand",
		Arguments: []cli.Argument{
			cli.Argument{Name: "subcommand", Parameter: HelpSubcommandParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleHelp(ctx)
		},
	}

	CSetup.Subcommands["init"] = &cli.Subcommand{
		Description: "Initialize a new workspace",
		Arguments: []cli.Argument{
			cli.Argument{Name: "workspace", Parameter: ccommon.WorkspaceParameter},
		},
		RequiredParams: []cli.Parameter{ccommon.WorkspaceParameter},
		Flags:          []cli.Flag{ccommon.ReinitFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleInit(ctx)
		},
	}
	CSetup.Subcommands["dev-init"] = &cli.Subcommand{
		Description: "Initialize a new workspace for development in the current directory",
		Arguments: []cli.Argument{
			cli.Argument{Name: "workspace", Parameter: ccommon.WorkspaceParameter},
		},
		Flags: []cli.Flag{ccommon.DownloadDepsFlag, ccommon.NoSetupFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDevInit(ctx)
		},
	}
	CSetup.Subcommands["git-clone"] = &cli.Subcommand{
		Description: "Clone a git repository into the workspace",
		Arguments: []cli.Argument{
			cli.Argument{Name: "url", Parameter: URLParameter},
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		RequiredParams: []cli.Parameter{URLParameter},
		Flags: []cli.Flag{
			ccommon.DownloadDepsFlag,
			ccommon.SubmoduleFlag,
			ccommon.NoSetupFlag,
			cli.Flag{Short: "", Long: "branch", Parameter: BranchParameter},
			cli.Flag{Short: "", Long: "revision", Parameter: RevisionParameter},
			cli.Flag{Short: "", Long: "no-track", Parameter: NoTrackParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleGitClone(ctx)
		},
	}
	CSetup.Subcommands["declare-git-source"] = &cli.Subcommand{
		Description: "Add git source information to the workspace without downloading",
		Arguments: []cli.Argument{
			cli.Argument{Name: "url", Parameter: URLParameter},
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		RequiredParams: []cli.Parameter{URLParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleDeclareGitSource(ctx)
		},
	}
	CSetup.Subcommands["declare-local-source"] = &cli.Subcommand{
		Description: "Add local source information to the workspace",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
			cli.Argument{Name: "sourcepath", Parameter: SourcePathParameter},
		},
		RequiredParams: []cli.Parameter{SourceParameter, SourcePathParameter},
		Flags:          []cli.Flag{SourcePathFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDeclareLocalSource(ctx)
		},
	}
	CSetup.Subcommands["download"] = &cli.Subcommand{
		Description: "Download missing sources",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		Flags: []cli.Flag{ccommon.DownloadDepsFlag, ccommon.NoSetupFlag, ccommon.SubmoduleFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDownload(ctx)
		},
	}
	CSetup.Subcommands["load-defaults"] = &cli.Subcommand{
		Description: "Load default configuration for a source from its csetup.yml",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		RequiredParams: []cli.Parameter{SourceParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleLoadDefaults(ctx)
		},
	}
	CSetup.Subcommands["add-dependency"] = &cli.Subcommand{
		Description: "Add a dependency to a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			cli.Argument{Name: "dependency", Parameter: DependencyParameter},
		},
		RequiredParams: []cli.Parameter{TargetParameter, DependencyParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddDependency(ctx)
		},
	}
	CSetup.Subcommands["remove-dependency"] = &cli.Subcommand{
		Description: "Remove a dependency from a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			cli.Argument{Name: "dependency", Parameter: DependencyParameter},
		},
		RequiredParams: []cli.Parameter{TargetParameter, DependencyParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveDependency(ctx)
		},
	}
	CSetup.Subcommands["add-testing-dependency"] = &cli.Subcommand{
		Description: "Add a testing dependency to a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			cli.Argument{Name: "dependency", Parameter: DependencyParameter},
		},
		RequiredParams: []cli.Parameter{TargetParameter, DependencyParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddTestingDependency(ctx)
		},
	}
	CSetup.Subcommands["remove-testing-dependency"] = &cli.Subcommand{
		Description: "Remove a testing dependency from a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			cli.Argument{Name: "dependency", Parameter: DependencyParameter},
		},
		RequiredParams: []cli.Parameter{TargetParameter, DependencyParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveTestingDependency(ctx)
		},
	}
	CSetup.Subcommands["enable-testing"] = &cli.Subcommand{
		Description: "Enable testing for a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
		},
		RequiredParams: []cli.Parameter{TargetParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleEnableTesting(ctx)
		},
	}
	CSetup.Subcommands["disable-testing"] = &cli.Subcommand{
		Description: "Disable testing for a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
		},
		RequiredParams: []cli.Parameter{TargetParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleDisableTesting(ctx)
		},
	}
	CSetup.Subcommands["remove-source"] = &cli.Subcommand{
		Description: "Remove a source from the workspace",
		Arguments: []cli.Argument{
			ccommon.SourceArg,
		},
		RequiredParams: []cli.Parameter{ccommon.SourceParameter},
		Flags:          []cli.Flag{ccommon.DeleteFlag, ccommon.SourceFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveSource(ctx)
		},
	}
	CSetup.Subcommands["remove-target"] = &cli.Subcommand{
		Description: "Remove a target from the workspace",
		Arguments: []cli.Argument{
			ccommon.TargetArg,
		},
		RequiredParams: []cli.Parameter{ccommon.TargetParameter},
		Flags:          []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveTarget(ctx)
		},
	}
	CSetup.Subcommands["remove-project"] = &cli.Subcommand{
		Description: "Remove a source and all its associated targets from the workspace",
		Arguments: []cli.Argument{
			ccommon.SourceArg,
		},
		RequiredParams: []cli.Parameter{ccommon.SourceParameter},
		Flags:          []cli.Flag{ccommon.DeleteFlag, ccommon.SourceFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveProject(ctx)
		},
	}
	CSetup.Subcommands["tidy"] = &cli.Subcommand{
		Description: "Delete source folders that are not in the sources list",
		Flags:       []cli.Flag{ccommon.DryRunFlag},
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
		RequiredParams: []cli.Parameter{ccommon.SourceParameter},
		Flags: []cli.Flag{
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
		RequiredParams: []cli.Parameter{ccommon.CxxVersionParameter},
		Flags:          []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleSetCXXVersion(ctx)
		},
	}
	CSetup.Subcommands["enable-staging"] = &cli.Subcommand{
		Description: "Enable staging for a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			ccommon.TargetArg,
		},
		RequiredParams: []cli.Parameter{ccommon.TargetParameter},
		Flags:          []cli.Flag{ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleEnableStaging(ctx)
		},
	}
	CSetup.Subcommands["disable-staging"] = &cli.Subcommand{
		Description: "Disable staging for a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			ccommon.TargetArg,
		},
		RequiredParams: []cli.Parameter{ccommon.TargetParameter},
		Flags:          []cli.Flag{ccommon.TargetFlag},
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
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleDropFiles(ctx)
		},
	}
	CSetup.Subcommands["get-args"] = &cli.Subcommand{
		Description: "Get build arguments for a target",
		Arguments: []cli.Argument{
			cli.Argument{Name: "target", Parameter: TargetParameter},
			ccommon.TargetArg,
		},
		RequiredParams: []cli.Parameter{ccommon.TargetParameter},
		Flags:          []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleGetArgs(ctx)
		},
	}
	CSetup.Subcommands["detect-toolchains"] = &cli.Subcommand{
		Description: "Detect system toolchains",
		Flags:       []cli.Flag{ccommon.DebugFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleDetectToolchains(ctx)
		},
	}
	CSetup.Subcommands["add-config"] = &cli.Subcommand{
		Description:    "Add a build configuration",
		RequiredParams: []cli.Parameter{ccommon.ConfigParameter},
		Flags:          []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleAddConfig(ctx)
		},
	}
	CSetup.Subcommands["remove-config"] = &cli.Subcommand{
		Description:    "Remove a build configuration",
		RequiredParams: []cli.Parameter{ccommon.ConfigParameter},
		Flags:          []cli.Flag{ccommon.ConfigFlag},
		Exec: func(ctx context.Context, args []string) error {
			return handleRemoveConfig(ctx)
		},
	}
	CSetup.Subcommands["pin"] = &cli.Subcommand{
		Description: "Pin a git source (HEAD by default, or a specific revision)",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		Flags: []cli.Flag{
			cli.Flag{Short: "", Long: "revision", Parameter: RevisionParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handlePin(ctx)
		},
	}
	CSetup.Subcommands["track"] = &cli.Subcommand{
		Description: "Track a branch for a git source",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
			cli.Argument{Name: "branch", Parameter: BranchParameter},
		},
		RequiredParams: []cli.Parameter{SourceParameter, BranchParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleTrack(ctx)
		},
	}
	CSetup.Subcommands["untrack"] = &cli.Subcommand{
		Description: "Stop tracking a branch for a git source",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		RequiredParams: []cli.Parameter{SourceParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleUntrack(ctx)
		},
	}
	CSetup.Subcommands["unpin"] = &cli.Subcommand{
		Description: "Unpin a git source (clear revision)",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		RequiredParams: []cli.Parameter{SourceParameter},
		Exec: func(ctx context.Context, args []string) error {
			return handleUnpin(ctx)
		},
	}
	CSetup.Subcommands["status"] = &cli.Subcommand{
		Description: "Show status for sources (clean/modified, HEAD, expected revision)",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleStatus(ctx)
		},
	}
	CSetup.Subcommands["versions"] = &cli.Subcommand{
		Description: "List pinned/unpinned versions for sources",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleVersions(ctx)
		},
	}
	CSetup.Subcommands["update"] = &cli.Subcommand{
		Description: "Update a git source (pull latest or move to revision)",
		Arguments: []cli.Argument{
			cli.Argument{Name: "source", Parameter: SourceParameter},
		},
		RequiredParams: []cli.Parameter{SourceParameter},
		Flags: []cli.Flag{
			cli.Flag{Short: "", Long: "revision", Parameter: RevisionParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleUpdate(ctx)
		},
	}
}
