package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
	"strings"
)

func handleGetArgs(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	targetNameVal := cli.GetOptionalString(ctx, TargetParameter)
	targetName := ""
	if targetNameVal != nil {
		targetName = *targetNameVal
	}
	toolchainVal := cli.GetOptionalString(ctx, ccommon.ToolchainParameter)
	toolchain := ""
	if toolchainVal != nil {
		toolchain = *toolchainVal
	}
	if toolchain == "" {
		toolchain = "default"
	}
	buildTypeVal := cli.GetOptionalStringList(ctx, ccommon.ConfigParameter)
	buildType := "Debug"
	if buildTypeVal != nil && len(*buildTypeVal) > 0 {
		buildType = (*buildTypeVal)[0]
	}

	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	bp := ccommon.TargetBuildParameters{
		Toolchain: toolchain,
		BuildType: buildType,
	}

	filteredArgs, err := ws.GetBuildArgs(ctx, targetName, bp)
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(filteredArgs, " "))
	return nil
}
