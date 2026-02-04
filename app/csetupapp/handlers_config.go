package csetupapp

import (
	"context"
	"fmt"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleSetCXXVersion(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	version := cli.GetString(ctx, ccommon.PCxxVersion)
	if version == nil {
		return fmt.Errorf("no CXX version provided")
	}
	target := cli.GetString(ctx, PTarget)
	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetCXXVersion(ctx, *version, target)
}

func handleEnableStaging(ctx context.Context) error {
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

	return ws.SetStaging(ctx, target, true)
}

func handleDisableStaging(ctx context.Context) error {
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

	return ws.SetStaging(ctx, target, false)
}

func handleAddConfig(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	configNameVal := cli.GetStringList(ctx, ccommon.PConfig)
	configName := ""
	if configNameVal != nil && len(*configNameVal) > 0 {
		configName = (*configNameVal)[0]
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.AddConfiguration(ctx, configName)
}

func handleRemoveConfig(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	configNameVal := cli.GetStringList(ctx, ccommon.PConfig)
	configName := ""
	if configNameVal != nil && len(*configNameVal) > 0 {
		configName = (*configNameVal)[0]
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.RemoveConfiguration(ctx, configName)
}
