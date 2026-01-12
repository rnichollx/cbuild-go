package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
)

func handleDetectToolchains(ctx context.Context, workspacePath string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: csetup detect-toolchains")
	}
	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.DetectToolchains(ctx)
}
