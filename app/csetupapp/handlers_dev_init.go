package csetupapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleDevInit(ctx context.Context) error {
	workspaceNameVal := cli.GetOptionalPath(ctx, ccommon.WorkspaceParameter)
	workspaceName := "dev-workspace"
	if workspaceNameVal != nil && *workspaceNameVal != "" {
		workspaceName = *workspaceNameVal
	}
	noSetupRaw := cli.GetOptionalBool(ctx, ccommon.NoSetupParameter)
	noSetup := noSetupRaw != nil && *noSetupRaw

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	currentDirAbs, err := filepath.Abs(currentDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path of current directory: %w", err)
	}

	workspacePathAbs, err := filepath.Abs(workspaceName)
	if err != nil {
		return fmt.Errorf("error getting absolute path of workspace: %w", err)
	}

	sourceName := filepath.Base(currentDirAbs)
	sourceLocalPath := currentDirAbs
	if relPath, err := filepath.Rel(workspacePathAbs, currentDirAbs); err == nil && relPath != "" {
		sourceLocalPath = relPath
	}

	ws := &ccommon.WorkspaceContext{
		WorkspacePath: workspaceName,
	}

	err = ws.Init(ctx, false)
	if err != nil {
		return fmt.Errorf("error initializing workspace: %w", err)
	}

	// Reload workspace to make sure everything is set up correctly
	err = ws.Load(ctx, workspaceName)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	if ws.Config.Sources == nil {
		ws.Config.Sources = make(map[string]*ccommon.CodeSource)
	}
	ws.Config.Sources[sourceName] = &ccommon.CodeSource{
		Local: sourceLocalPath,
	}

	if ws.Config.Targets == nil {
		ws.Config.Targets = make(map[string]*ccommon.TargetConfiguration)
	}
	// Initial target for the source
	ws.Config.Targets[sourceName] = &ccommon.TargetConfiguration{
		Source: sourceName,
	}

	err = ws.Save(ctx)
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Println("Detecting toolchains...")
	err = ws.DetectToolchains(ctx)
	if err != nil {
		return fmt.Errorf("error detecting toolchains: %w", err)
	}

	if noSetup {
		fmt.Println("Skipping setup as requested by --no-setup.")
	} else {
		loadDefaultsCtx := ctx

		// dev-init defaults to auto-downloading suggested dependencies unless explicitly disabled.
		downloadDeps := cli.GetBoolOr(ctx, ccommon.DownloadDepsParameter, cli.PBool(true))
		loadDefaultsCtx, err = cli.SetBool(loadDefaultsCtx, ccommon.DownloadDepsParameter, *downloadDeps)
		if err != nil {
			return fmt.Errorf("error setting default download behavior: %w", err)
		}
		loadDefaultsCtx = ccommon.WithApplyProjectDefaults(loadDefaultsCtx, true)

		// Process csetup defaults unless explicitly disabled.
		err = ws.LoadDefaults(loadDefaultsCtx, sourceName)
		if err != nil {
			return fmt.Errorf("error loading defaults from CSetup.yml: %w", err)
		}
	}

	fmt.Printf("Initialized dev workspace in %s with source %s from %s\n", workspaceName, sourceName, currentDirAbs)
	return nil
}
