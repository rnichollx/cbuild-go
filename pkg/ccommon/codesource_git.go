package ccommon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

type GitSource struct {
	Repository string  `yaml:"repository"`
	Revision   *string `yaml:"revision,omitempty"`
	Branch     *string `yaml:"branch,omitempty"`
}

func (ws *WorkspaceContext) GetFromGit(ctx context.Context, name string, source GitSource) error {
	destDir := filepath.Join(ws.WorkspacePath, "sources", name)
	useSubmoduleRaw := cli.GetOptionalBool(ctx, SubmoduleParameter)
	useSubmodule := useSubmoduleRaw != nil && *useSubmoduleRaw

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
	} else if source.Branch != nil {
		fmt.Printf("Checking out branch '%s' for '%s'...\n", *source.Branch, name)
		checkoutCmd := exec.CommandContext(ctx, "git", "checkout", *source.Branch)
		checkoutCmd.Dir = destDir
		checkoutCmd.Stdout = os.Stdout
		checkoutCmd.Stderr = os.Stderr
		err = checkoutCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to checkout branch '%s' for '%s': %w", *source.Branch, name, err)
		}
	}

	return nil
}

func (ws *WorkspaceContext) GitStatus(ctx context.Context, repoPath string) (dirty bool, head string, err error) {
	headOut, err := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to get git HEAD: %w", err)
	}
	head = strings.TrimSpace(string(headOut))

	statusOut, err := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--porcelain").Output()
	if err != nil {
		return false, head, fmt.Errorf("failed to get git status: %w", err)
	}
	dirty = strings.TrimSpace(string(statusOut)) != ""
	return dirty, head, nil
}

func (ws *WorkspaceContext) GitResolveRevision(ctx context.Context, repoPath string, revision string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", revision).Output()
	if err != nil {
		return "", fmt.Errorf("failed to resolve revision %q: %w", revision, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (ws *WorkspaceContext) GitRevisionOnBranch(ctx context.Context, repoPath string, revision string, branch string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "merge-base", "--is-ancestor", revision, branch)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
		return false, nil
	}
	return false, fmt.Errorf("failed to check if %q is on branch %q: %w", revision, branch, err)
}

func (ws *WorkspaceContext) GitFetch(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "fetch", "--all", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (ws *WorkspaceContext) GitCheckout(ctx context.Context, repoPath string, revision string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", revision)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (ws *WorkspaceContext) GitPull(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull", "--ff-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (ws *WorkspaceContext) GitCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		return "", nil
	}
	return branch, nil
}
