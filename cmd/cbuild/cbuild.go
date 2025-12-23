package main

import (
	"cbuild-go/pkg/ccommon"
	"cbuild-go/pkg/cli"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	if workspacePath == "" {
		workspacePath = "."
	}
	toolchain := cli.GetString(ctx, cli.FlagKey(ccommon.FlagToolchain))
	if toolchain == "" {
		toolchain = "all"
	}
	dryRun := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagDryRun))

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	configs := []string{}
	if buildConfig == "" {
		configs = ws.Configurations
	} else {
		configs = strings.Split(buildConfig, ",")
	}

	if len(configs) == 0 {
		err = ws.Clean(toolchain, "", dryRun)
		if err != nil {
			return fmt.Errorf("error cleaning workspace: %w", err)
		}
	} else {
		for _, cfg := range configs {
			cfg = strings.TrimSpace(cfg)
			err = ws.Clean(toolchain, cfg, dryRun)
			if err != nil {
				return fmt.Errorf("error cleaning workspace for toolchain %s, config %s: %w", toolchain, cfg, err)
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

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
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
		configs = ws.Configurations
	} else {
		configs = strings.Split(buildConfig, ",")
	}

	for _, tc := range toolchains {
		for _, cfg := range configs {
			cfg = strings.TrimSpace(cfg)
			fmt.Printf("Building with toolchain: %s, config: %s\n", tc, cfg)

			bp := ccommon.BuildParameters{
				Toolchain: tc,
				BuildType: cfg,
				DryRun:    dryRun,
			}

			if command == "build-deps" {
				err = ws.BuildDependencies(targetName, bp)
			} else if targetName != "" {
				err = ws.BuildTarget(targetName, bp)
			} else {
				err = ws.Build(bp)
			}

			if err != nil {
				return fmt.Errorf("error building workspace for toolchain %s, config %s: %w", tc, cfg, err)
			}
		}
	}

	fmt.Println("Build completed successfully")
	return nil
}
