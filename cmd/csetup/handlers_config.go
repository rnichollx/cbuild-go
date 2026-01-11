package main

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
)

func handleSetCXXVersion(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup set-cxx-version <version> [<source>]")
	}

	version := args[0]
	var source string
	if len(args) > 1 {
		source = args[1]
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetCXXVersion(ctx, version, source)
}

func handleEnableStaging(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup enable-staging <source>")
	}

	source := args[0]

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetStaging(ctx, source, true)
}

func handleDisableStaging(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup disable-staging <source>")
	}

	source := args[0]

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetStaging(ctx, source, false)
}

func handleAddConfig(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup add-config <configname>")
	}

	configName := args[0]

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.AddConfiguration(ctx, configName)
}

func handleRemoveConfig(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup remove-config <configname>")
	}

	configName := args[0]

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.RemoveConfiguration(ctx, configName)
}
