package ccommon

import (
	"context"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

const (
	FlagWorkspace        cli.ParameterKey = "workspace"
	FlagConfig           cli.ParameterKey = "config"
	FlagTarget           cli.ParameterKey = "target"
	FlagToolchain        cli.ParameterKey = "toolchain"
	FlagDryRun           cli.ParameterKey = "dry-run"
	FlagReinit           cli.ParameterKey = "reinit"
	FlagDownload         cli.ParameterKey = "download-deps"
	FlagSubmodule        cli.ParameterKey = "submodule"
	FlagDelete           cli.ParameterKey = "delete"
	FlagNoSetup          cli.ParameterKey = "no-setup"
	FlagHelp             cli.ParameterKey = "help"
	FlagSource           cli.ParameterKey = "source"
	FlagDebug            cli.ParameterKey = "debug"
	FlagOverwrite        cli.ParameterKey = "overwrite"
	FlagProjectType      cli.ParameterKey = "project-type"
	FlagCMakePackageName cli.ParameterKey = "cmake-package-name"
)

var (
	PWorkspace    = cli.NewParameter(FlagWorkspace, cli.ParameterTypePath, nil, "Path to the workspace directory.")
	WorkspaceFlag = cli.NewStringFlag("w", "workspace", PWorkspace)

	PConfig    = cli.NewParameter(FlagConfig, cli.ParameterTypeStringList, nil, "Build configuration to use (e.g., Debug, Release), comma separated")
	ConfigFlag = cli.NewStringFlag("c", "config", PConfig)

	PTarget    = cli.NewParameter(FlagTarget, cli.ParameterTypeString, nil, "Specific target to build/modify.")
	TargetArg  = cli.NewStringArgument("target", PTarget)
	TargetFlag = cli.NewStringFlag("t", "target", PTarget)

	PSource    = cli.NewParameter(FlagSource, cli.ParameterTypeString, nil, "Specific source to modify.")
	SourceArg  = cli.NewStringArgument("source", PSource)
	SourceFlag = cli.NewStringFlag("s", "source", PSource)

	PToolchain    = cli.NewParameter(FlagToolchain, cli.ParameterTypeString, nil, "Select toolchain to use.")
	ToolchainFlag = cli.NewStringFlag("T", "toolchain", PToolchain)

	PDryRun    = cli.NewParameter(FlagDryRun, cli.ParameterTypeBool, cli.PBool(false), "Show commands without executing them.")
	DryRunFlag = cli.NewBoolFlag("d", "dry-run", PDryRun)

	PReinit    = cli.NewParameter(FlagReinit, cli.ParameterTypeBool, cli.PBool(false), "Reinitialize the workspace.")
	ReinitFlag = cli.NewBoolFlag("", "reinit", PReinit)

	PDownloadDeps    = cli.NewParameter(FlagDownload, cli.ParameterTypeBool, cli.PBool(false), "Download dependencies during clone.")
	DownloadDepsFlag = cli.NewBoolFlag("", "download-deps", PDownloadDeps)

	PSubmodule    = cli.NewParameter(FlagSubmodule, cli.ParameterTypeBool, cli.PBool(false), "Add as a git submodule instead of cloning.")
	SubmoduleFlag = cli.NewBoolFlag("", "submodule", PSubmodule)

	PDelete    = cli.NewParameter(FlagDelete, cli.ParameterTypeBool, cli.PBool(false), "Delete files when removing source.")
	DeleteFlag = cli.NewBoolFlag("X", "delete", PDelete)

	PNoSetup    = cli.NewParameter(FlagNoSetup, cli.ParameterTypeBool, cli.PBool(false), "Don't run setup after downloading or cloning.")
	NoSetupFlag = cli.NewBoolFlag("", "no-setup", PNoSetup)

	PHelp    = cli.NewParameter(FlagHelp, cli.ParameterTypeBool, cli.PBool(false), "Show this help message.")
	HelpFlag = cli.NewBoolFlag("h", "help", PHelp)

	PDebug    = cli.NewParameter(FlagDebug, cli.ParameterTypeBool, cli.PBool(false), "Show debug information.")
	DebugFlag = cli.NewBoolFlag("", "debug", PDebug)

	POverwrite    = cli.NewParameter(FlagOverwrite, cli.ParameterTypeBool, cli.PBool(false), "Overwrite target if it exists.")
	OverwriteFlag = cli.NewBoolFlag("", "overwrite", POverwrite)

	PProjectType    = cli.NewParameter(FlagProjectType, cli.ParameterTypeString, nil, "The project type (e.g., CMake).")
	ProjectTypeFlag = cli.NewStringFlag("p", "project-type", PProjectType)

	PCMakePackageName    = cli.NewParameter(FlagCMakePackageName, cli.ParameterTypeString, nil, "The CMake package name.")
	CMakePackageNameFlag = cli.NewStringFlag("", "cmake-package-name", PCMakePackageName)

	PCxxVersion   = cli.NewParameter("version", cli.ParameterTypeString, nil, "C++ version number.")
	CxxVersionArg = cli.NewStringArgument("version", PCxxVersion)
)

func IsDebug(ctx context.Context) bool {
	val := cli.GetBool(ctx, PDebug)
	return val != nil && *val
}
