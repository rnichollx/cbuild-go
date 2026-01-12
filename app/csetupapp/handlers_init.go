package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
	"strings"
)

func handleInit(ctx context.Context, workspacePath string, args []string) error {
	reinit := cli.GetBool(ctx, cli.FlagKey(ccommon.FlagReinit))
	workspaceName := ""

	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") && workspaceName == "" {
			workspaceName = arg
		}
	}

	if workspaceName == "" || len(args) > 1 {
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
