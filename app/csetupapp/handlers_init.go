package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleInit(ctx context.Context) error {
	reinitRaw := cli.GetBool(ctx, ccommon.PReinit)
	reinit := reinitRaw != nil && *reinitRaw
	workspaceNameVal := cli.GetPath(ctx, PPath)
	workspaceName := ""
	if workspaceNameVal != nil {
		workspaceName = *workspaceNameVal
	}

	if workspaceName == "" {
		return fmt.Errorf("usage: csetup init <workspace name> [--reinit]")
	}

	ws := &ccommon.WorkspaceContext{
		WorkspacePath: workspaceName,
	}

	err := ws.Init(ctx, reinit)
	if err != nil {
		return err
	}

	fmt.Printf("Initialized empty workspace in %s\n", workspaceName)
	return nil
}
