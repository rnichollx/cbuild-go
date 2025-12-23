package ccommon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GitSource struct {
	Repository string  `yaml:"repository"`
	Revision   *string `yaml:"revision,omitempty"`
}

func (ws *Workspace) GetFromGit(name string, source GitSource) error {
	destDir := filepath.Join(ws.WorkspacePath, "sources", name)
	fmt.Printf("Downloading '%s' from '%s'...\n", name, source.Repository)

	args := []string{"clone", source.Repository, destDir}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to download '%s': %w", name, err)
	}

	if source.Revision != nil {
		fmt.Printf("Checking out revision '%s' for '%s'...\n", *source.Revision, name)
		checkoutCmd := exec.Command("git", "checkout", *source.Revision)
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
