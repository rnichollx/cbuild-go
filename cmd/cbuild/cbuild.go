package main

import (
	"cbuild-go/pkg/ccommon"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var buildConfig string
	var workspacePath string
	var targetName string
	var toolchain string
	var dryRun bool

	flag.StringVar(&buildConfig, "c", "Debug", "build configuration to use (e.g., Debug, Release)")
	flag.StringVar(&buildConfig, "config", "Debug", "build configuration to use (e.g., Debug, Release)")

	flag.StringVar(&workspacePath, "w", ".", "path to the workspace directory")
	flag.StringVar(&workspacePath, "workspace", ".", "path to the workspace directory")

	flag.StringVar(&targetName, "t", "", "specific target to build")
	flag.StringVar(&targetName, "target", "", "specific target to build")

	flag.StringVar(&toolchain, "T", "all", "toolchain to use (use 'all' to build all toolchains)")
	flag.StringVar(&toolchain, "toolchain", "all", "toolchain to use (use 'all' to build all toolchains)")

	flag.BoolVar(&dryRun, "dry-run", false, "show commands without executing them")
	flag.BoolVar(&dryRun, "dry_run", false, "show commands without executing them")

	flag.Parse()

	command := "build"
	args := flag.Args()
	if len(args) > 0 {
		if args[0] == "clean" || args[0] == "build" {
			command = args[0]
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
			os.Exit(1)
		}
	}

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if command == "clean" {
		err = ws.Clean(toolchain, dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning workspace: %v\n", err)
			os.Exit(1)
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

	for _, tc := range toolchains {
		fmt.Printf("Building with toolchain: %s, config: %s\n", tc, buildConfig)

		bp := ccommon.BuildParameters{
			Toolchain: tc,
			BuildType: buildConfig,
			DryRun:    dryRun,
		}

		if targetName != "" {
			err = ws.BuildTarget(targetName, bp)
		} else {
			err = ws.Build(bp)
		}

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error building workspace for toolchain %s: %v\n", tc, err)
			os.Exit(1)
		}
	}

	fmt.Println("Build completed successfully")
}
