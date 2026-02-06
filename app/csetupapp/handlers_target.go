package csetupapp

import (
	"context"
	"fmt"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleNewTarget(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)

	source := cli.GetOptionalString(ctx, ccommon.SourceParameter)
	if source == nil {
		return fmt.Errorf("source is required")
	}

	target := cli.GetOptionalString(ctx, ccommon.TargetParameter)

	targetName := *source
	if target != nil && *target != "" {
		targetName = *target
	}

	overwriteVal := cli.GetOptionalBool(ctx, ccommon.OverwriteParameter)
	overwrite := overwriteVal != nil && *overwriteVal

	projectType := cli.GetOptionalString(ctx, ccommon.ProjectTypeParameter)
	cmakePackageName := cli.GetOptionalString(ctx, ccommon.CMakePackageNameParameter)

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
	targetVal := cli.GetOptionalString(ctx, TargetParameter)
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
	targetVal := cli.GetOptionalString(ctx, TargetParameter)
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
