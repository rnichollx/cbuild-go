package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleAddDependency(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	targetVal, _ := cli.GetString(ctx, PTargetReq)
	target := ""
	if targetVal != nil {
		target = *targetVal
	}
	depnameVal, _ := cli.GetString(ctx, PDependency)
	depname := ""
	if depnameVal != nil {
		depname = *depnameVal
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	err = ws.AddDependency(ctx, target, depname)
	if err != nil {
		return err
	}

	fmt.Printf("Added dependency %s to %s\n", depname, target)
	return nil
}

func handleRemoveDependency(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	targetVal, _ := cli.GetString(ctx, PTargetReq)
	target := ""
	if targetVal != nil {
		target = *targetVal
	}
	depnameVal, _ := cli.GetString(ctx, PDependency)
	depname := ""
	if depnameVal != nil {
		depname = *depnameVal
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	err = ws.RemoveDependency(ctx, target, depname)
	if err != nil {
		return err
	}

	fmt.Printf("Removed dependency %s from %s\n", depname, target)
	return nil
}
