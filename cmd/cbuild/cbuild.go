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

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	toolchains := []string{}
	if toolchain == "all" {
		toolchainDir := filepath.Join(ws.WorkspacePath, "toolchains")
		files, err := os.ReadDir(toolchainDir)
		if err != nil {
			// If toolchains directory doesn't exist, use default
			fmt.Println("No toolchains directory found, using 'default'")
			toolchains = append(toolchains, "default")
		} else {
			for _, file := range files {
				if file.IsDir() {
					toolchains = append(toolchains, file.Name())
				}
			}
			if len(toolchains) == 0 {
				fmt.Println("No toolchains found in toolchains directory, using 'default'")
				toolchains = append(toolchains, "default")
			}
		}
	} else {
		toolchains = append(toolchains, toolchain)
	}

	for _, tc := range toolchains {
		fmt.Printf("Building with toolchain: %s, config: %s\n", tc, buildConfig)

		tcPath := ""
		if tc != "default" {
			tcPath = filepath.Join(ws.WorkspacePath, "toolchains", tc)
		}

		bp := ccommon.BuildParameters{
			Toolchain:     tc,
			ToolchainPath: tcPath,
			BuildType:     buildConfig,
			DryRun:        dryRun,
		}

		if targetName != "" {
			err = ws.BuildTarget(targetName, bp)
		} else {
			err = ws.Build(bp)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building workspace for toolchain %s: %v\n", tc, err)
			os.Exit(1)
		}
	}

	fmt.Println("Build completed successfully")
}
