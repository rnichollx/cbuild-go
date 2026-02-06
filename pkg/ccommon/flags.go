package ccommon

import (
	"context"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

const (
	WorkspaceKey        cli.ParameterKey = "workspace"
	ConfigKey           cli.ParameterKey = "config"
	TargetKey           cli.ParameterKey = "target"
	ToolchainKey        cli.ParameterKey = "toolchain"
	DryRunKey           cli.ParameterKey = "dry-run"
	ReinitKey           cli.ParameterKey = "reinit"
	DownloadDepsKey     cli.ParameterKey = "download-deps"
	SubmoduleKey        cli.ParameterKey = "submodule"
	DeleteKey           cli.ParameterKey = "delete"
	NoSetupKey          cli.ParameterKey = "no-setup"
	HelpKey             cli.ParameterKey = "help"
	SourceKey           cli.ParameterKey = "source"
	DebugKey            cli.ParameterKey = "debug"
	OverwriteKey        cli.ParameterKey = "overwrite"
	ProjectTypeKey      cli.ParameterKey = "project-type"
	CMakePackageNameKey cli.ParameterKey = "cmake-package-name"
)

var (
	WorkspaceParameter = cli.Parameter{Key: WorkspaceKey, Type: cli.ParameterTypePath, DefaultValue: nil, Description: "Path to the workspace directory."}
	WorkspaceFlag      = cli.Flag{Short: "w", Long: "workspace", Parameter: WorkspaceParameter}

	ConfigParameter = cli.Parameter{Key: ConfigKey, Type: cli.ParameterTypeStringList, DefaultValue: nil, Description: "Build configuration to use (e.g., Debug, Release), comma separated"}
	ConfigFlag      = cli.Flag{Short: "c", Long: "config", Parameter: ConfigParameter}

	TargetParameter = cli.Parameter{Key: TargetKey, Type: cli.ParameterTypeString, DefaultValue: nil, Description: "Specific target to build/modify."}
	TargetArg       = cli.Argument{Name: "target", Parameter: TargetParameter}
	TargetFlag      = cli.Flag{Short: "t", Long: "target", Parameter: TargetParameter}

	SourceParameter = cli.Parameter{Key: SourceKey, Type: cli.ParameterTypeString, DefaultValue: nil, Description: "Specific source to modify."}
	SourceArg       = cli.Argument{Name: "source", Parameter: SourceParameter}
	SourceFlag      = cli.Flag{Short: "s", Long: "source", Parameter: SourceParameter}

	ToolchainParameter = cli.Parameter{Key: ToolchainKey, Type: cli.ParameterTypeString, DefaultValue: nil, Description: "Select toolchain to use."}
	ToolchainFlag      = cli.Flag{Short: "T", Long: "toolchain", Parameter: ToolchainParameter}

	DryRunParameter = cli.Parameter{Key: DryRunKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Show commands without executing them."}
	DryRunFlag      = cli.Flag{Short: "d", Long: "dry-run", Parameter: DryRunParameter}

	ReinitParameter = cli.Parameter{Key: ReinitKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Reinitialize the workspace."}
	ReinitFlag      = cli.Flag{Short: "", Long: "reinit", Parameter: ReinitParameter}

	DownloadDepsParameter = cli.Parameter{Key: DownloadDepsKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Download dependencies during clone."}
	DownloadDepsFlag      = cli.Flag{Short: "", Long: "download-deps", Parameter: DownloadDepsParameter}

	SubmoduleParameter = cli.Parameter{Key: SubmoduleKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Add as a git submodule instead of cloning."}
	SubmoduleFlag      = cli.Flag{Short: "", Long: "submodule", Parameter: SubmoduleParameter}

	DeleteParameter = cli.Parameter{Key: DeleteKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Delete files when removing source."}
	DeleteFlag      = cli.Flag{Short: "X", Long: "delete", Parameter: DeleteParameter}

	NoSetupParameter = cli.Parameter{Key: NoSetupKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Don't run setup after downloading or cloning."}
	NoSetupFlag      = cli.Flag{Short: "", Long: "no-setup", Parameter: NoSetupParameter}

	HelpParameter = cli.Parameter{Key: HelpKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Show this help message."}
	HelpFlag      = cli.Flag{Short: "h", Long: "help", Parameter: HelpParameter}

	DebugParameter = cli.Parameter{Key: DebugKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Show debug information."}
	DebugFlag      = cli.Flag{Short: "", Long: "debug", Parameter: DebugParameter}

	OverwriteParameter = cli.Parameter{Key: OverwriteKey, Type: cli.ParameterTypeBool, DefaultValue: cli.PBool(false), Description: "Overwrite target if it exists."}
	OverwriteFlag      = cli.Flag{Short: "", Long: "overwrite", Parameter: OverwriteParameter}

	ProjectTypeParameter = cli.Parameter{Key: ProjectTypeKey, Type: cli.ParameterTypeString, DefaultValue: nil, Description: "The project type (e.g., CMake)."}
	ProjectTypeFlag      = cli.Flag{Short: "p", Long: "project-type", Parameter: ProjectTypeParameter}

	CMakePackageNameParameter = cli.Parameter{Key: CMakePackageNameKey, Type: cli.ParameterTypeString, DefaultValue: nil, Description: "The CMake package name."}
	CMakePackageNameFlag      = cli.Flag{Short: "", Long: "cmake-package-name", Parameter: CMakePackageNameParameter}

	CxxVersionParameter = cli.Parameter{Key: "version", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "C++ version number."}
	CxxVersionArg       = cli.Argument{Name: "version", Parameter: CxxVersionParameter}
)

func IsDebug(ctx context.Context) bool {
	val := cli.GetOptionalBool(ctx, DebugParameter)
	return val != nil && *val
}
