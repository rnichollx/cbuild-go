package csetupapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func ensureGitSource(ws *ccommon.WorkspaceContext, sourceName string) (*ccommon.CodeSource, error) {
	source, ok := ws.Config.Sources[sourceName]
	if !ok {
		return nil, fmt.Errorf("source %s not found in workspace configuration", sourceName)
	}
	if source.Git == nil {
		return nil, fmt.Errorf("source %s is not a git source", sourceName)
	}
	return source, nil
}

func isExternallyManagedSource(source *ccommon.CodeSource) bool {
	if source == nil {
		return false
	}
	return source.Local != ""
}

func resolveSourcePath(ws *ccommon.WorkspaceContext, sourceName string) (string, error) {
	sourcePath, err := ws.GetSourcePath(sourceName)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return "", fmt.Errorf("source %s not found on disk at %s", sourceName, sourcePath)
	}
	return sourcePath, nil
}

func handlePin(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSource)
	if sourceVal == nil || *sourceVal == "" {
		return fmt.Errorf("usage: csetup pin <source>")
	}
	branchVal, _ := cli.GetString(ctx, PBranch)
	revisionVal, _ := cli.GetString(ctx, PRevision)
	updateVal, _ := cli.GetBool(ctx, PUpdate)
	noBranchVal, _ := cli.GetBool(ctx, PNoBranch)
	update := updateVal != nil && *updateVal
	noBranch := noBranchVal != nil && *noBranchVal

	if noBranch && branchVal != nil && *branchVal != "" {
		return fmt.Errorf("use either --no-branch or --branch, not both")
	}
	if branchVal != nil && *branchVal != "" && revisionVal != nil && *revisionVal != "" {
		return fmt.Errorf("use either --branch or --revision, not both")
	}
	if update && revisionVal != nil && *revisionVal != "" {
		return fmt.Errorf("use either --update or --revision, not both")
	}

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	source, err := ensureGitSource(ws, *sourceVal)
	if err != nil {
		return err
	}
	if isExternallyManagedSource(source) {
		return fmt.Errorf("source %s is externally managed; update is only supported for sources in the workspace sources list", *sourceVal)
	}

	sourcePath, err := resolveSourcePath(ws, *sourceVal)
	if err != nil {
		return err
	}

	if noBranch {
		source.Git.Branch = nil
	}

	if branchVal != nil && *branchVal != "" {
		branch := *branchVal
		source.Git.Branch = &branch

		if err := ws.GitFetch(ctx, sourcePath); err != nil {
			return err
		}
		if update {
			if err := ws.GitCheckout(ctx, sourcePath, branch); err != nil {
				return err
			}
			if err := ws.GitPull(ctx, sourcePath); err != nil {
				return err
			}
			_, head, err := ws.GitStatus(ctx, sourcePath)
			if err != nil {
				return err
			}
			if source.Git.Revision == nil || *source.Git.Revision == "" {
				source.Git.Revision = &head
			}
			if err := ws.Save(ctx); err != nil {
				return err
			}
			fmt.Printf("Pinned %s to branch %s (updated to %s)\n", *sourceVal, branch, head)
			return nil
		}

		_, head, err := ws.GitStatus(ctx, sourcePath)
		if err != nil {
			return err
		}

		if source.Git.Revision == nil || *source.Git.Revision == "" {
			source.Git.Revision = &head
		}

		if err := ws.Save(ctx); err != nil {
			return err
		}
		fmt.Printf("Pinned %s to branch %s\n", *sourceVal, branch)
		return nil
	}

	if revisionVal != nil && *revisionVal != "" {
		revision := *revisionVal
		if err := ws.GitFetch(ctx, sourcePath); err != nil {
			return err
		}
		resolved, err := ws.GitResolveRevision(ctx, sourcePath, revision)
		if err != nil {
			return err
		}
		if err := ws.GitCheckout(ctx, sourcePath, resolved); err != nil {
			return err
		}
		source.Git.Revision = &resolved

		if source.Git.Branch != nil && *source.Git.Branch != "" {
			onBranch, err := ws.GitRevisionOnBranch(ctx, sourcePath, resolved, *source.Git.Branch)
			if err != nil {
				return err
			}
			if !onBranch {
				source.Git.Branch = nil
			}
		}

		if err := ws.Save(ctx); err != nil {
			return err
		}
		fmt.Printf("Pinned %s to %s\n", *sourceVal, resolved)
		return nil
	}

	if update {
		if source.Git.Branch == nil || *source.Git.Branch == "" {
			if source.Git.Revision != nil && *source.Git.Revision != "" {
				return fmt.Errorf("cannot --update without a tracked branch (source is pinned to %s)", *source.Git.Revision)
			}
			return fmt.Errorf("cannot --update without a tracked branch")
		}
		if err := ws.GitFetch(ctx, sourcePath); err != nil {
			return err
		}
		if err := ws.GitCheckout(ctx, sourcePath, *source.Git.Branch); err != nil {
			return err
		}
		if err := ws.GitPull(ctx, sourcePath); err != nil {
			return err
		}
		_, head, err := ws.GitStatus(ctx, sourcePath)
		if err != nil {
			return err
		}
		if source.Git.Revision == nil || *source.Git.Revision == "" {
			source.Git.Revision = &head
		}
		if err := ws.Save(ctx); err != nil {
			return err
		}
		fmt.Printf("Pinned %s to branch %s (updated to %s)\n", *sourceVal, *source.Git.Branch, head)
		return nil
	}

	_, head, err := ws.GitStatus(ctx, sourcePath)
	if err != nil {
		return err
	}

	source.Git.Revision = &head
	if err := ws.Save(ctx); err != nil {
		return err
	}

	fmt.Printf("Pinned %s to %s\n", *sourceVal, head)
	return nil
}

func handleUnpin(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSource)
	if sourceVal == nil || *sourceVal == "" {
		return fmt.Errorf("usage: csetup unpin <source>")
	}

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	source, err := ensureGitSource(ws, *sourceVal)
	if err != nil {
		return err
	}
	if isExternallyManagedSource(source) {
		return fmt.Errorf("source %s is externally managed; unpin is only supported for sources in the workspace sources list", *sourceVal)
	}

	source.Git.Revision = nil
	if err := ws.Save(ctx); err != nil {
		return err
	}

	fmt.Printf("Unpinned %s\n", *sourceVal)
	return nil
}

func handleUpdate(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSource)
	revisionVal, _ := cli.GetString(ctx, PRevision)
	if sourceVal == nil || *sourceVal == "" {
		return fmt.Errorf("usage: csetup update <source> [--revision <revision>]")
	}

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	source, err := ensureGitSource(ws, *sourceVal)
	if err != nil {
		return err
	}

	sourcePath, err := resolveSourcePath(ws, *sourceVal)
	if err != nil {
		return err
	}

	if err := ws.GitFetch(ctx, sourcePath); err != nil {
		return err
	}

	if revisionVal != nil && *revisionVal != "" {
		resolved, err := ws.GitResolveRevision(ctx, sourcePath, *revisionVal)
		if err != nil {
			return err
		}
		if err := ws.GitCheckout(ctx, sourcePath, resolved); err != nil {
			return err
		}
		source.Git.Revision = &resolved
		if source.Git.Branch != nil && *source.Git.Branch != "" {
			onBranch, err := ws.GitRevisionOnBranch(ctx, sourcePath, resolved, *source.Git.Branch)
			if err != nil {
				return err
			}
			if !onBranch {
				source.Git.Branch = nil
			}
		}
		if err := ws.Save(ctx); err != nil {
			return err
		}
		fmt.Printf("Updated %s to %s\n", *sourceVal, resolved)
		return nil
	}

	if source.Git.Branch != nil && *source.Git.Branch != "" {
		if err := ws.GitCheckout(ctx, sourcePath, *source.Git.Branch); err != nil {
			return err
		}
		if err := ws.GitPull(ctx, sourcePath); err != nil {
			return err
		}
		_, head, err := ws.GitStatus(ctx, sourcePath)
		if err != nil {
			return err
		}
		source.Git.Revision = &head
		if err := ws.Save(ctx); err != nil {
			return err
		}
		fmt.Printf("Updated %s to %s\n", *sourceVal, head)
		return nil
	}

	if source.Git.Revision != nil && *source.Git.Revision != "" {
		return fmt.Errorf("source %s is pinned to %s but has no tracked branch; use --revision to change revision or set a branch", *sourceVal, *source.Git.Revision)
	}

	if err := ws.GitPull(ctx, sourcePath); err != nil {
		return err
	}
	fmt.Printf("Updated %s to latest\n", *sourceVal)
	return nil
}

func handleStatus(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal, _ := cli.GetString(ctx, PSource)

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	sources := []string{}
	if sourceVal != nil && *sourceVal != "" {
		sources = append(sources, *sourceVal)
	} else {
		for name := range ws.Config.Sources {
			sources = append(sources, name)
		}
	}

	for _, name := range sources {
		source, ok := ws.Config.Sources[name]
		if !ok {
			return fmt.Errorf("source %s not found in workspace configuration", name)
		}
		if isExternallyManagedSource(source) {
			fmt.Printf("%s EXTERNAL (local source)\n", name)
			continue
		}

		sourcePath, err := ws.GetSourcePath(name)
		if err != nil {
			return err
		}

		if _, err := os.Stat(sourcePath); err != nil {
			fmt.Printf("%s MISSING (%s)\n", name, sourcePath)
			continue
		}

		if source.Git == nil {
			fmt.Printf("%s LOCAL (%s)\n", name, sourcePath)
			continue
		}

		dirty, head, err := ws.GitStatus(ctx, sourcePath)
		if err != nil {
			return err
		}

		expected := ""
		if source.Git.Revision != nil {
			expected = *source.Git.Revision
		}

		resolvedExpected := ""
		if expected != "" {
			resolvedExpected, _ = ws.GitResolveRevision(ctx, sourcePath, expected)
		}

		modified := dirty
		if expected != "" && resolvedExpected != "" && resolvedExpected != head {
			modified = true
		}

		status := "OK"
		if modified {
			status = "MODIFIED"
		}

		relPath := sourcePath
		if rel, err := filepath.Rel(workspacePath, sourcePath); err == nil {
			relPath = rel
		}

		if expected != "" {
			fmt.Printf("%s %s (path=%s head=%s expected=%s)\n", name, status, relPath, head, expected)
		} else {
			fmt.Printf("%s %s (path=%s head=%s)\n", name, status, relPath, head)
		}
	}

	return nil
}
