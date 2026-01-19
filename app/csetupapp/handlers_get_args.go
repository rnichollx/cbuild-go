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
	targetNameVal, _ := cli.GetString(ctx, PTargetReq)
	targetName := ""
	if targetNameVal != nil {
		targetName = *targetNameVal
	}
	toolchainVal, _ := cli.GetString(ctx, ccommon.PToolchain)
	toolchain := ""
	if toolchainVal != nil {
		toolchain = *toolchainVal
	}
	if toolchain == "" {
		toolchain = "default"
	}
	buildTypeVal, _ := cli.GetStringList(ctx, ccommon.PConfig)
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
