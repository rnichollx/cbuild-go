package ccommon

import (
	"context"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

const (
	FlagWorkspace FlagKey = "workspace"
	FlagConfig    FlagKey = "config"
	FlagTarget    FlagKey = "target"
	FlagToolchain FlagKey = "toolchain"
	FlagDryRun    FlagKey = "dry-run"
	FlagReinit    FlagKey = "reinit"
	FlagDownload  FlagKey = "download-deps"
	FlagSubmodule FlagKey = "submodule"
	FlagDelete    FlagKey = "delete"
	FlagNoSetup   FlagKey = "no-setup"
	FlagHelp      FlagKey = "help"
	FlagSource    FlagKey = "source"
	FlagDebug     FlagKey = "debug"
)

type FlagKey string

var (
	PWorkspace    = cli.NewParameter(cli.ParameterKey(FlagWorkspace), cli.ParameterTypePath, nil, "Path to the workspace directory.", false)
	WorkspaceFlag = cli.NewStringFlag("w", "workspace", PWorkspace)

	PConfig    = cli.NewParameter(cli.ParameterKey(FlagConfig), cli.ParameterTypeStringList, nil, "Build configuration to use (e.g., Debug, Release), comma separated", false)
	ConfigFlag = cli.NewStringFlag("c", "config", PConfig)

	PTarget    = cli.NewParameter(cli.ParameterKey(FlagTarget), cli.ParameterTypeString, nil, "Specific target to build/modify.", false)
	TargetArg  = cli.NewStringArgument("target", PTarget)
	TargetFlag = cli.NewStringFlag("t", "target", PTarget)

	PSource    = cli.NewParameter(cli.ParameterKey(FlagSource), cli.ParameterTypeString, nil, "Specific source to modify.", false)
	SourceArg  = cli.NewStringArgument("source", PSource)
	SourceFlag = cli.NewStringFlag("s", "source", PSource)

	PToolchain    = cli.NewParameter(cli.ParameterKey(FlagToolchain), cli.ParameterTypeString, nil, "Select toolchain to use.", false)
	ToolchainFlag = cli.NewStringFlag("T", "toolchain", PToolchain)

	PDryRun    = cli.NewParameter(cli.ParameterKey(FlagDryRun), cli.ParameterTypeBool, cli.PBool(false), "Show commands without executing them.", false)
	DryRunFlag = cli.NewBoolFlag("d", "dry-run", PDryRun)

	PReinit    = cli.NewParameter(cli.ParameterKey(FlagReinit), cli.ParameterTypeBool, cli.PBool(false), "Reinitialize the workspace.", false)
	ReinitFlag = cli.NewBoolFlag("", "reinit", PReinit)

	PDownloadDeps    = cli.NewParameter(cli.ParameterKey(FlagDownload), cli.ParameterTypeBool, cli.PBool(false), "Download dependencies during clone.", false)
	DownloadDepsFlag = cli.NewBoolFlag("", "download-deps", PDownloadDeps)

	PSubmodule    = cli.NewParameter(cli.ParameterKey(FlagSubmodule), cli.ParameterTypeBool, cli.PBool(false), "Add as a git submodule instead of cloning.", false)
	SubmoduleFlag = cli.NewBoolFlag("", "submodule", PSubmodule)

	PDelete    = cli.NewParameter(cli.ParameterKey(FlagDelete), cli.ParameterTypeBool, cli.PBool(false), "Delete files when removing source.", false)
	DeleteFlag = cli.NewBoolFlag("X", "delete", PDelete)

	PNoSetup    = cli.NewParameter(cli.ParameterKey(FlagNoSetup), cli.ParameterTypeBool, cli.PBool(false), "Don't run setup after downloading or cloning.", false)
	NoSetupFlag = cli.NewBoolFlag("", "no-setup", PNoSetup)

	PHelp    = cli.NewParameter(cli.ParameterKey(FlagHelp), cli.ParameterTypeBool, cli.PBool(false), "Show this help message.", false)
	HelpFlag = cli.NewBoolFlag("h", "help", PHelp)

	PDebug    = cli.NewParameter(cli.ParameterKey(FlagDebug), cli.ParameterTypeBool, cli.PBool(false), "Show debug information.", false)
	DebugFlag = cli.NewBoolFlag("", "debug", PDebug)

	PCxxVersion   = cli.NewParameter("version", cli.ParameterTypeString, nil, "C++ version number.", true)
	CxxVersionArg = cli.NewStringArgument("version", PCxxVersion)
)

func IsDebug(ctx context.Context) bool {
	val, _ := cli.GetBool(ctx, PDebug)
	return val != nil && *val
}
