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
	flags := []cli.Flag{
		ccommon.ConfigFlag,
		ccommon.WorkspaceFlag,
		ccommon.TargetFlag,
		ccommon.ToolchainFlag,
		ccommon.DryRunFlag,
	}

	ctx, args, err := cli.ParseFlags(context.Background(), os.Args[1:], flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

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

	command := "build"
	if len(args) > 0 {
		if args[0] == "clean" || args[0] == "build" || args[0] == "build-deps" {
			command = args[0]
			if command == "build-deps" {
				if len(args) < 2 {
					fmt.Fprintf(os.Stderr, "Usage: cbuild build-deps <sourcename> [-T toolchain] [-c config]\n")
					os.Exit(1)
				}
				targetName = args[1]
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
			os.Exit(1)
		}
	}

	ws := &ccommon.Workspace{}
	err = ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if command == "clean" {
		configs := []string{}
		if buildConfig == "" {
			configs = ws.Configurations
		} else {
			configs = strings.Split(buildConfig, ",")
		}

		if len(configs) == 0 {
			// If no configs in workspace and none specified, just clean the toolchain directory
			err = ws.Clean(toolchain, "", dryRun)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error cleaning workspace: %v\n", err)
				os.Exit(1)
			}
		} else {
			for _, cfg := range configs {
				cfg = strings.TrimSpace(cfg)
				err = ws.Clean(toolchain, cfg, dryRun)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error cleaning workspace for toolchain %s, config %s: %v\n", toolchain, cfg, err)
					os.Exit(1)
				}
			}
		}
		fmt.Println("Clean completed successfully")
		return
	}

	toolchains := []string{}
	if toolchain == "all" {
		toolchainDir := filepath.Join(ws.WorkspacePath, "toolchains")
		files, err := os.ReadDir(toolchainDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading toolchains directory: %v\n", err)
			os.Exit(1)
		} else {
			for _, file := range files {
				if file.IsDir() {
					toolchains = append(toolchains, file.Name())
				}
			}
			if len(toolchains) == 0 {
				fmt.Fprintf(os.Stderr, "No toolchains found in toolchains directory\n")
				os.Exit(1)
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
				_, _ = fmt.Fprintf(os.Stderr, "Error building workspace for toolchain %s, config %s: %v\n", tc, cfg, err)
				os.Exit(1)
			}
		}
	}

	fmt.Println("Build completed successfully")
}
