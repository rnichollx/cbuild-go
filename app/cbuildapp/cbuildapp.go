package cbuildapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

var (
	HelpSubcommandParameter = cli.Parameter{Key: "help-subcommand", Type: cli.ParameterTypeString, DefaultValue: nil, Description: "subcommand to show help for"}
)

func init() {
	CBuild.Subcommands["help"] = &cli.Subcommand{
		Description: "Show help for cbuild or a specific subcommand",
		Arguments: []cli.Argument{
			{Name: "subcommand", Parameter: HelpSubcommandParameter},
		},
		Exec: func(ctx context.Context, args []string) error {
			return handleHelp(ctx)
		},
	}

	CBuild.Subcommands["build"] = &cli.Subcommand{
		Description: "Build the project",
		Flags:       []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Arguments:   []cli.Argument{ccommon.TargetArg},
		Exec: func(ctx context.Context, args []string) error {
			return runBuild(ctx, "build", args)
		},
	}

	CBuild.Subcommands["clean"] = &cli.Subcommand{
		Description: "Clean build artifacts",
		Flags:       []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Arguments:   []cli.Argument{ccommon.TargetArg},
		Exec: func(ctx context.Context, args []string) error {
			return runClean(ctx, args)
		},
	}

	CBuild.Subcommands["build-deps"] = &cli.Subcommand{
		Description: "Build dependencies for a target",
		Arguments: []cli.Argument{
			ccommon.TargetArg,
		},
		RequiredParams: []cli.Parameter{ccommon.TargetParameter},
		Flags:          []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Exec: func(ctx context.Context, args []string) error {
			targetNameVal := cli.GetOptionalString(ctx, ccommon.TargetParameter)
			if targetNameVal == nil || *targetNameVal == "" {
				return fmt.Errorf("usage: cbuild build-deps <targetname>")
			}
			return runBuild(ctx, "build-deps", args)
		},
	}
	CBuild.Subcommands["test"] = &cli.Subcommand{
		Description: "Build and run tests",
		Flags:       []cli.Flag{ccommon.ConfigFlag, ccommon.ToolchainFlag, ccommon.TargetFlag},
		Arguments:   []cli.Argument{ccommon.TargetArg},
		Exec: func(ctx context.Context, args []string) error {
			return runTest(ctx, args)
		},
	}
}

func runClean(ctx context.Context, args []string) error {
	buildConfigRaw := cli.GetOptionalStringList(ctx, ccommon.ConfigParameter)
	workspacePathRaw := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter)
	targetFlagRaw := cli.GetOptionalString(ctx, ccommon.TargetParameter)
	workspacePath := ""
	if workspacePathRaw != nil {
		workspacePath = *workspacePathRaw
	}
	if workspacePath == "" {
		workspacePath = "."
	}

	dryRunRaw := cli.GetOptionalBool(ctx, ccommon.DryRunParameter)
	dryRun := dryRunRaw != nil && *dryRunRaw

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchainFlagRaw := cli.GetOptionalString(ctx, ccommon.ToolchainParameter)
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
	buildConfigRaw := cli.GetOptionalStringList(ctx, ccommon.ConfigParameter)
	workspacePathRaw := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter)
	workspacePath := ""
	if workspacePathRaw != nil {
		workspacePath = *workspacePathRaw
	}
	if workspacePath == "" {
		workspacePath = "."
	}
	targetNameRaw := cli.GetOptionalString(ctx, ccommon.TargetParameter)
	targetName := ""
	if targetNameRaw != nil {
		targetName = *targetNameRaw
	}
	toolchainRaw := cli.GetOptionalString(ctx, ccommon.ToolchainParameter)
	toolchain := ""
	if toolchainRaw != nil {
		toolchain = *toolchainRaw
	}
	if toolchain == "" {
		toolchain = "all"
	}
	dryRunRaw := cli.GetOptionalBool(ctx, ccommon.DryRunParameter)
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

func runTest(ctx context.Context, args []string) error {
	buildConfigRaw := cli.GetOptionalStringList(ctx, ccommon.ConfigParameter)
	workspacePathRaw := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter)
	targetFlagRaw := cli.GetOptionalString(ctx, ccommon.TargetParameter)
	workspacePath := ""
	if workspacePathRaw != nil {
		workspacePath = *workspacePathRaw
	}
	if workspacePath == "" {
		workspacePath = "."
	}

	dryRunRaw := cli.GetOptionalBool(ctx, ccommon.DryRunParameter)
	dryRun := dryRunRaw != nil && *dryRunRaw

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	toolchainFlagRaw := cli.GetOptionalString(ctx, ccommon.ToolchainParameter)
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
	if targetFlagRaw != nil && len(*targetFlagRaw) > 0 {
		targets = strings.Split(*targetFlagRaw, ",")
	} else {
		targets = ws.ListTargets(ctx)
	}

	allResults := []ccommon.TestResult{}

	if targetFlagRaw != nil && len(*targetFlagRaw) > 0 {
		// Build only specific targets if requested, but Build() builds all
		// so we use BuildTarget for each.
		for _, toolchainName := range toolchainNames {
			for _, buildType := range configs {
				bp := ccommon.TargetBuildParameters{
					Toolchain: toolchainName,
					BuildType: buildType,
					DryRun:    dryRun,
					Testing:   true,
				}

				for _, targetName := range targets {
					fmt.Printf("Building %s for %s (%s) for testing\n", targetName, toolchainName, buildType)
					err = ws.BuildTarget(ctx, targetName, bp)
					if err != nil {
						return fmt.Errorf("failed to build %s: %w", targetName, err)
					}
				}
			}
		}
	} else {
		// Build all first
		for _, toolchainName := range toolchainNames {
			for _, buildType := range configs {
				bp := ccommon.TargetBuildParameters{
					Toolchain: toolchainName,
					BuildType: buildType,
					DryRun:    dryRun,
					Testing:   true,
				}

				fmt.Printf("Building for %s (%s) for testing\n", toolchainName, buildType)
				err = ws.Build(ctx, bp)
				if err != nil {
					return fmt.Errorf("failed to build for testing: %w", err)
				}
			}
		}
	}

	// Then run tests
	for _, toolchainName := range toolchainNames {
		for _, buildType := range configs {
			bp := ccommon.TargetBuildParameters{
				Toolchain: toolchainName,
				BuildType: buildType,
				DryRun:    dryRun,
				Testing:   true,
			}
			results, err := ws.RunTests(ctx, bp, targets)
			if err != nil {
				return err
			}
			allResults = append(allResults, results...)
		}
	}

	// Generate reports
	if !dryRun {
		timestamp := time.Now().Format("20060102-150405")
		reportDir := filepath.Join(workspacePath, "reports", fmt.Sprintf("testing-%s", timestamp))
		err = os.MkdirAll(reportDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create report directory: %w", err)
		}

		summaryPath := filepath.Join(reportDir, "summary.md")
		summaryFile, err := os.Create(summaryPath)
		if err != nil {
			return fmt.Errorf("failed to create summary file: %w", err)
		}
		defer summaryFile.Close()

		fmt.Fprintf(summaryFile, "# Test Summary - %s\n\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(summaryFile, "| Target | Toolchain | Config | Status | Report |\n")
		fmt.Fprintf(summaryFile, "|--------|-----------|--------|--------|--------|\n")

		for _, res := range allResults {
			status := "PASS"
			if res.NoTests {
				status = "NO_CTEST"
			} else if !res.Passed {
				status = "FAIL"
			}
			reportName := fmt.Sprintf("report-%s-%s-%s.log", res.Target, res.Toolchain, res.Config)
			fmt.Fprintf(summaryFile, "| %s | %s | %s | %s | [%s](%s) |\n", res.Target, res.Toolchain, res.Config, status, reportName, reportName)

			reportPath := filepath.Join(reportDir, reportName)
			err = os.WriteFile(reportPath, res.Output, 0644)
			if err != nil {
				return fmt.Errorf("failed to write report file %s: %w", reportName, err)
			}
		}

		if len(allResults) > 0 {
			fmt.Println()
			fmt.Println("Test Summary:")
			fmt.Printf("%-20s %-20s %-15s %-10s\n", "Target", "Toolchain", "Config", "Status")
			fmt.Println(strings.Repeat("-", 68))
			passedCount := 0
			noTestsCount := 0
			for _, res := range allResults {
				status := "PASS"
				if res.NoTests {
					status = "NO_CTEST"
					noTestsCount++
					passedCount++
				} else if !res.Passed {
					status = "FAIL"
				} else {
					passedCount++
				}
				fmt.Printf("%-20s %-20s %-15s %-10s\n", res.Target, res.Toolchain, res.Config, status)
			}
			fmt.Println(strings.Repeat("-", 68))
			if noTestsCount > 0 {
				fmt.Printf("Result: %d/%d passed (%d with no tests)\n", passedCount, len(allResults), noTestsCount)
			} else {
				fmt.Printf("Result: %d/%d passed\n", passedCount, len(allResults))
			}
			fmt.Println()
		}

		fmt.Printf("Detailed test reports saved in %s\n", reportDir)
	}

	return nil
}
