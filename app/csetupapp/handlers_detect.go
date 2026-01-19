package csetupapp

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
)

func handleDetectToolchains(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	ws := &ccommon.WorkspaceContext{}
	err := ws.Load(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	return ws.DetectToolchains(ctx)
}
