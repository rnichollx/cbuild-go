package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
)

func handleAddDependency(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: csetup add-dependency <source> <depname>")
	}

	source := args[0]
	depname := args[1]

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

func handleRemoveDependency(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: csetup remove-dependency <source> <depname>")
	}

	source := args[0]
	depname := args[1]

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
