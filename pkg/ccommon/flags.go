package ccommon

import (
	"cbuild-go/pkg/cli"
)

const (
	FlagWorkspace FlagKey = "workspace"
	FlagConfig    FlagKey = "config"
	FlagTarget    FlagKey = "target"
	FlagToolchain FlagKey = "toolchain"
	FlagDryRun    FlagKey = "dry-run"
	FlagReinit    FlagKey = "reinit"
	FlagDownload  FlagKey = "download-deps"
	FlagDelete    FlagKey = "delete"
)

type FlagKey string

var (
	WorkspaceFlag = cli.NewStringFlag("w", "workspace", cli.FlagKey(FlagWorkspace), "path to the workspace directory")

	ConfigFlag = cli.NewStringFlag("c", "config", cli.FlagKey(FlagConfig), "build configuration to use (e.g., Debug, Release), comma separated")

	TargetFlag = cli.NewStringFlag("t", "target", cli.FlagKey(FlagTarget), "specific target to build")

	ToolchainFlag = cli.NewStringFlag("T", "toolchain", cli.FlagKey(FlagToolchain), "toolchain to use")

	DryRunFlag = cli.NewBoolFlag("", "dry-run", cli.FlagKey(FlagDryRun), "show commands without executing them")

	ReinitFlag = cli.NewBoolFlag("", "reinit", cli.FlagKey(FlagReinit), "reinitialize the workspace")

	DownloadDepsFlag = cli.NewBoolFlag("", "download-deps", cli.FlagKey(FlagDownload), "download dependencies during clone")

	DeleteFlag = cli.NewBoolFlag("D", "delete", cli.FlagKey(FlagDelete), "delete files when removing source")
)
