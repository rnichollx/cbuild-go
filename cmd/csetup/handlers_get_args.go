package main

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
	"strings"
)

func handleGetArgs(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
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
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	target, err := ws.GetTarget(targetName)
	if err != nil {
		return err
	}

	bp := ccommon.TargetBuildParameters{
		Toolchain: toolchain,
		BuildType: buildType,
	}

	fullArgs, err := target.CMakeConfigureArgs(context.Background(), ws, bp)
	if err != nil {
		return fmt.Errorf("error getting cmake args: %w", err)
	}

	filteredArgs := []string{}
	for i := 0; i < len(fullArgs); i++ {
		arg := fullArgs[i]
		if arg == "-S" || arg == "-B" {
			i++ // skip the next argument too (the path)
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	fmt.Println(strings.Join(filteredArgs, " "))
	return nil
}
