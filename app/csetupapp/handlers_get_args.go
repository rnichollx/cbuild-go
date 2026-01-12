package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
	"strings"
)

func handleGetArgs(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: csetup get-args <target> [-T|--toolchain <toolchain>] [-c|--config <type>]")
	}

	targetName := args[0]
	toolchain := cli.GetString(ctx, cli.FlagKey(ccommon.FlagToolchain))
	if toolchain == "" {
		toolchain = "default"
	}
	buildType := cli.GetString(ctx, cli.FlagKey(ccommon.FlagConfig))
	if buildType == "" {
		buildType = "Debug"
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
