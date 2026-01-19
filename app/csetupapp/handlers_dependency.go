package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleAddDependency(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSourceReq)
	source := ""
	if sourceVal != nil {
		source = *sourceVal
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

	err = ws.AddDependency(ctx, source, depname)
	if err != nil {
		return err
	}

	fmt.Printf("Added dependency %s to %s\n", depname, source)
	return nil
}

func handleRemoveDependency(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSourceReq)
	source := ""
	if sourceVal != nil {
		source = *sourceVal
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

	err = ws.RemoveDependency(ctx, source, depname)
	if err != nil {
		return err
	}

	fmt.Printf("Removed dependency %s from %s\n", depname, source)
	return nil
}
