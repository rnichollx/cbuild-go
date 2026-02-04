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
	workspaceNameVal, _ := cli.GetPath(ctx, PPath)
	workspaceName := "workspace"
	if workspaceNameVal != nil && *workspaceNameVal != "" {
		workspaceName = *workspaceNameVal
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	currentDirAbs, err := filepath.Abs(currentDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path of current directory: %w", err)
	}

	sourceName := filepath.Base(currentDirAbs)

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
		Local: currentDirAbs,
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

	// Now we need to process CSetup.yml if it exists
	err = ws.LoadDefaults(ctx, sourceName)
	if err != nil {
		return fmt.Errorf("error loading defaults from CSetup.yml: %w", err)
	}

	err = ws.Save(ctx)
	if err != nil {
		return fmt.Errorf("error saving workspace after processing CSetup.yml: %w", err)
	}

	fmt.Printf("Initialized dev workspace in %s with source %s from %s\n", workspaceName, sourceName, currentDirAbs)
	return nil
}
