package csetupapp

import (
	"context"
	"fmt"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleSetCXXVersion(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	version, err := cli.GetString(ctx, ccommon.PCxxVersion)
	if err != nil {
		return err
	}
	if version == nil {
		return fmt.Errorf("no CXX version provided")
	}
	target, err := cli.GetString(ctx, PTargetReq)
	if err != nil {
		return fmt.Errorf("getString: %w", err)
	}
	ws := &ccommon.WorkspaceContext{}
	err = ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetCXXVersion(ctx, *version, target)
}

func handleEnableStaging(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSource)
	source := ""
	if sourceVal != nil {
		source = *sourceVal
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetStaging(ctx, source, true)
}

func handleDisableStaging(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSource)
	source := ""
	if sourceVal != nil {
		source = *sourceVal
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.SetStaging(ctx, source, false)
}

func handleAddConfig(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	configNameVal, _ := cli.GetStringList(ctx, ccommon.PConfig)
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
	configNameVal, _ := cli.GetStringList(ctx, ccommon.PConfig)
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
