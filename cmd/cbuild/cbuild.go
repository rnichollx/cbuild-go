package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func main() {
	runner := &cli.Runner{
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

	runner.Subcommands["build"] = &cli.Subcommand{
		Description:  "Build the project",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			return runBuild(ctx, "build", args)
		},
	}

	runner.Subcommands["clean"] = &cli.Subcommand{
		Description:  "Clean build artifacts",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			return runClean(ctx, args)
		},
	}

	runner.Subcommands["build-deps"] = &cli.Subcommand{
		Description:  "Build dependencies for a source",
		AllowArgs:    true,
		AcceptsFlags: []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: cbuild build-deps <sourcename>")
			}
			return runBuild(ctx, "build-deps", args)
		},
	}

	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runClean(ctx context.Context, args []string) error {
	buildConfig := cli.GetString(ctx, cli.FlagKey(ccommon.FlagConfig))
	workspacePath := cli.GetString(ctx, cli.FlagKey(ccommon.FlagWorkspace))
	targetFlag := cli.GetString(ctx, cli.FlagKey(ccommon.FlagTarget))
	if workspacePath == "" {
		workspacePath = "."
	}

	dryRun := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDryRun))

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchainFlag := cli.GetString(ctx, cli.FlagKey(ccommon.FlagToolchain))

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
	if buildConfig == "" {
		configs = ws.Config.Configurations
	} else {
		configs = strings.Split(buildConfig, ",")
	}

	var targets []string
	if len(targetFlag) == 0 {
		targets = ws.ListTargets(ctx)
	} else {
		targets = strings.Split(targetFlag, ",")
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
	buildConfig := cli.GetString(ctx, cli.FlagKey(ccommon.FlagConfig))
	workspacePath := cli.GetString(ctx, cli.FlagKey(ccommon.FlagWorkspace))
	if workspacePath == "" {
		workspacePath = "."
	}
	targetName := cli.GetString(ctx, cli.FlagKey(ccommon.FlagTarget))
	toolchain := cli.GetString(ctx, cli.FlagKey(ccommon.FlagToolchain))
	if toolchain == "" {
		toolchain = "all"
	}
	dryRun := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDryRun))

	if command == "build-deps" {
		targetName = args[0]
	}

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
	if buildConfig == "" {
		configs = ws.Config.Configurations
	} else {
		configs = strings.Split(buildConfig, ",")
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
