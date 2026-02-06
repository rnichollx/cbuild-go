package csetupapp

import (
	"context"
	"fmt"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleInit(ctx context.Context) error {
	reinitRaw := cli.GetOptionalBool(ctx, ccommon.ReinitParameter)
	reinit := reinitRaw != nil && *reinitRaw
	workspacePath := cli.GetPath(ctx, ccommon.WorkspaceParameter)

	if workspacePath == "" {
		return fmt.Errorf("usage: csetup init <workspace path> [--reinit]")
	}

	ws := &ccommon.WorkspaceContext{
		WorkspacePath: workspacePath,
	}

	err := ws.Init(ctx, reinit)
	if err != nil {
		return err
	}

	fmt.Printf("Initialized empty workspace in %s\n", workspacePath)
	return nil
}
