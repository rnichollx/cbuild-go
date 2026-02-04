package csetupapp

import (
	"context"
	"fmt"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleNewTarget(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)

	source := cli.GetString(ctx, ccommon.PSource)
	if source == nil {
		return fmt.Errorf("source is required")
	}

	target := cli.GetString(ctx, ccommon.PTarget)

	targetName := *source
	if target != nil && *target != "" {
		targetName = *target
	}

	overwriteVal := cli.GetBool(ctx, ccommon.POverwrite)
	overwrite := overwriteVal != nil && *overwriteVal

	projectType := cli.GetString(ctx, ccommon.PProjectType)
	cmakePackageName := cli.GetString(ctx, ccommon.PCMakePackageName)

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	opts := ccommon.AddTargetOptions{
		SourceName:       *source,
		TargetName:       targetName,
		Overwrite:        overwrite,
		ProjectType:      projectType,
		CMakePackageName: cmakePackageName,
	}

	err = ws.AddTarget(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Added target %s (source: %s) to workspace\n", targetName, *source)
	return nil
}

func handleEnableTesting(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	targetVal := cli.GetString(ctx, PTarget)
	target := ""
	if targetVal != nil {
		target = *targetVal
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetTestingEnabled(ctx, target, true)
}

func handleDisableTesting(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	targetVal := cli.GetString(ctx, PTarget)
	target := ""
	if targetVal != nil {
		target = *targetVal
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetTestingEnabled(ctx, target, false)
}
