package ccommon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

type GitSource struct {
	Repository string  `yaml:"repository"`
	Revision   *string `yaml:"revision,omitempty"`
}

func (ws *WorkspaceContext) GetFromGit(ctx context.Context, name string, source GitSource) error {
	destDir := filepath.Join(ws.WorkspacePath, "sources", name)
	useSubmodule := cli.GetBool(ctx, cli.FlagKey(FlagSubmodule))

	if useSubmodule {
		fmt.Printf("Adding submodule '%s' from '%s'...\n", name, source.Repository)
	} else {
		fmt.Printf("Downloading '%s' from '%s'...\n", name, source.Repository)
	}

	var cmd *exec.Cmd
	if useSubmodule {
		relDestDir := filepath.Join("sources", name)
		cmd = exec.CommandContext(ctx, "git", "submodule", "add", source.Repository, relDestDir)
		cmd.Dir = ws.WorkspacePath
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", source.Repository, destDir)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if useSubmodule {
			return fmt.Errorf("failed to add submodule '%s': %w", name, err)
		}
		return fmt.Errorf("failed to download '%s': %w", name, err)
	}

	if source.Revision != nil {
		fmt.Printf("Checking out revision '%s' for '%s'...\n", *source.Revision, name)
		checkoutCmd := exec.CommandContext(ctx, "git", "checkout", *source.Revision)
		checkoutCmd.Dir = destDir
		checkoutCmd.Stdout = os.Stdout
		checkoutCmd.Stderr = os.Stderr
		err = checkoutCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to checkout revision '%s' for '%s': %w", *source.Revision, name, err)
		}
	}

	return nil
}
