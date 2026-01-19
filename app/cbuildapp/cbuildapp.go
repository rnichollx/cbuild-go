package cbuildapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

var CBuild = &cli.Runner{
	Name:        "cbuild",
	Description: "Build tool for cbuild projects",
	GlobalFlags: []cli.Flag{
		ccommon.WorkspaceFlag,
		ccommon.DryRunFlag,
		ccommon.HelpFlag,
	},
	Subcommands:   make(map[string]*cli.Subcommand),
	DefaultSubcmd: "build",
}

func init() {
	CBuild.Subcommands["build"] = &cli.Subcommand{
		Description:  "Build the project",
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Arguments:    []cli.Argument{ccommon.TargetArg},
		Exec: func(ctx context.Context, args []string) error {
			return runBuild(ctx, "build", args)
		},
	}

	CBuild.Subcommands["clean"] = &cli.Subcommand{
		Description:  "Clean build artifacts",
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Arguments:    []cli.Argument{ccommon.TargetArg},
		Exec: func(ctx context.Context, args []string) error {
			return runClean(ctx, args)
		},
	}

	CBuild.Subcommands["build-deps"] = &cli.Subcommand{
		Description: "Build dependencies for a target",
		Arguments: []cli.Argument{
			ccommon.TargetArg,
		},
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			targetNameVal, _ := cli.GetString(ctx, ccommon.PTarget)
			if targetNameVal == nil || *targetNameVal == "" {
				return fmt.Errorf("usage: cbuild build-deps <targetname>")
			}
			return runBuild(ctx, "build-deps", args)
		},
	}
}

func runClean(ctx context.Context, args []string) error {
	buildConfigRaw, _ := cli.GetStringList(ctx, ccommon.PConfig)
	workspacePathRaw, _ := cli.GetPath(ctx, ccommon.PWorkspace)
	targetFlagRaw, _ := cli.GetString(ctx, ccommon.PTarget)
	workspacePath := ""
	if workspacePathRaw != nil {
		workspacePath = *workspacePathRaw
	}
	if workspacePath == "" {
		workspacePath = "."
	}

	dryRunRaw, _ := cli.GetBool(ctx, ccommon.PDryRun)
	dryRun := dryRunRaw != nil && *dryRunRaw

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchainFlagRaw, _ := cli.GetString(ctx, ccommon.PToolchain)
	toolchainFlag := ""
	if toolchainFlagRaw != nil {
		toolchainFlag = *toolchainFlagRaw
	}

	var toolchainNames []string
	if len(toolchainFlag) == 0 {
		toolchainNames, err = ws.ListToolchains(ctx)
		if err != nil {
			return fmt.Errorf("error listing toolchains: %w", err)
		}
	} else {
		toolchainNames = strings.Split(toolchainFlag, ",")
	}

	configs := []string{}
	if buildConfigRaw == nil || len(*buildConfigRaw) == 0 {
		configs = ws.Config.Configurations
	} else {
		configs = *buildConfigRaw
	}

	var targets []string
	if targetFlagRaw == nil || len(*targetFlagRaw) == 0 {
		targets = ws.ListTargets(ctx)
	} else {
		targets = strings.Split(*targetFlagRaw, ",")
	}

	//fmt.Printf("Cleaning %d targets: %s\n", len(targets), targets)

	for _, target := range targets {
		for _, toolchain := range toolchainNames {
			for _, config := range configs {
				bp := ccommon.TargetBuildParameters{
					Toolchain: toolchain,
					BuildType: config,
					DryRun:    dryRun,
				}
				err = ws.CleanTarget(ctx, target, bp)
				if err != nil {
					return fmt.Errorf("error cleaning target %q: %w", target, err)
				}
			}
		}

	}

	fmt.Println("Clean completed successfully")
	return nil
}

func runBuild(ctx context.Context, command string, args []string) error {
	buildConfigRaw, _ := cli.GetStringList(ctx, ccommon.PConfig)
	workspacePathRaw, _ := cli.GetPath(ctx, ccommon.PWorkspace)
	workspacePath := ""
	if workspacePathRaw != nil {
		workspacePath = *workspacePathRaw
	}
	if workspacePath == "" {
		workspacePath = "."
	}
	targetNameRaw, _ := cli.GetString(ctx, ccommon.PTarget)
	targetName := ""
	if targetNameRaw != nil {
		targetName = *targetNameRaw
	}
	toolchainRaw, _ := cli.GetString(ctx, ccommon.PToolchain)
	toolchain := ""
	if toolchainRaw != nil {
		toolchain = *toolchainRaw
	}
	if toolchain == "" {
		toolchain = "all"
	}
	dryRunRaw, _ := cli.GetBool(ctx, ccommon.PDryRun)
	dryRun := dryRunRaw != nil && *dryRunRaw

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchains := []string{}
	if toolchain == "all" {
		toolchainDir := filepath.Join(ws.WorkspacePath, "toolchains")
		files, err := os.ReadDir(toolchainDir)
		if err != nil {
			return fmt.Errorf("error reading toolchains directory: %w", err)
		} else {
			for _, file := range files {
				if file.IsDir() {
					toolchains = append(toolchains, file.Name())
				}
			}
			if len(toolchains) == 0 {
				return fmt.Errorf("no toolchains found in toolchains directory")
			}
		}
	} else {
		toolchains = append(toolchains, toolchain)
	}

	configs := []string{}
	if buildConfigRaw == nil || len(*buildConfigRaw) == 0 {
		configs = ws.Config.Configurations
	} else {
		configs = *buildConfigRaw
	}

	for _, tc := range toolchains {
		for _, cfg := range configs {
			cfg = strings.TrimSpace(cfg)
			fmt.Printf("Building with toolchain: %s, config: %s\n", tc, cfg)

			bp := ccommon.TargetBuildParameters{
				Toolchain: tc,
				BuildType: cfg,
				DryRun:    dryRun,
			}

			if command == "build-deps" {
				err = ws.BuildDependencies(ctx, targetName, bp)
			} else if targetName != "" {
				err = ws.BuildTarget(ctx, targetName, bp)
			} else {
				err = ws.Build(ctx, bp)
			}

			if err != nil {
				return fmt.Errorf("error building workspace for toolchain %s, config %s: %w", tc, cfg, err)
			}
		}
	}

	fmt.Println("Build completed successfully")
	return nil
}
