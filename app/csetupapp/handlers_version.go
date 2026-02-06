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
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)
	revisionVal := cli.GetOptionalString(ctx, RevisionParameter)

	ws := &ccommon.WorkspaceContext{}
	if err := ws.Load(ctx, workspacePath); err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	targets := []string{}
	if sourceVal != nil && *sourceVal != "" {
		targets = append(targets, *sourceVal)
	} else {
		for name := range ws.Config.Sources {
			targets = append(targets, name)
		}
	}

	for _, name := range targets {
		source := ws.Config.Sources[name]
		if source == nil {
			return fmt.Errorf("source %s not found in workspace configuration", name)
		}
		if source.Git == nil {
			if sourceVal == nil || *sourceVal == "" {
				fmt.Printf("Skipping non-git source %s\n", name)
				continue
			}
			return fmt.Errorf("source %s is not a git source", name)
		}
		if isExternallyManagedSource(source) {
			if sourceVal == nil || *sourceVal == "" {
				fmt.Printf("Skipping external source %s\n", name)
				continue
			}
			return fmt.Errorf("source %s is externally managed; pin is only supported for sources in the workspace sources list", name)
		}

		sourcePath, err := resolveSourcePath(ws, name)
		if err != nil {
			return err
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
			fmt.Printf("Pinned %s to %s\n", name, resolved)
			continue
		}

		_, head, err := ws.GitStatus(ctx, sourcePath)
		if err != nil {
			return err
		}

		source.Git.Revision = &head
		if err := ws.Save(ctx); err != nil {
			return err
		}

		fmt.Printf("Pinned %s to %s\n", name, head)
	}

	return nil
}

func handleTrack(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)
	branchVal := cli.GetOptionalString(ctx, BranchParameter)
	if sourceVal == nil || *sourceVal == "" || branchVal == nil || *branchVal == "" {
		return fmt.Errorf("usage: csetup track <source> <branch>")
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
		return fmt.Errorf("source %s is externally managed; track is only supported for sources in the workspace sources list", *sourceVal)
	}

	branch := *branchVal
	source.Git.Branch = &branch
	if err := ws.Save(ctx); err != nil {
		return err
	}

	fmt.Printf("Tracking branch %s for %s\n", branch, *sourceVal)
	return nil
}

func handleUntrack(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)
	if sourceVal == nil || *sourceVal == "" {
		return fmt.Errorf("usage: csetup untrack <source>")
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
		return fmt.Errorf("source %s is externally managed; untrack is only supported for sources in the workspace sources list", *sourceVal)
	}

	source.Git.Branch = nil
	if err := ws.Save(ctx); err != nil {
		return err
	}

	fmt.Printf("Stopped tracking branch for %s\n", *sourceVal)
	return nil
}

func handleUnpin(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)
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
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)
	revisionVal := cli.GetOptionalString(ctx, RevisionParameter)
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
	if isExternallyManagedSource(source) {
		return fmt.Errorf("source %s is externally managed; update is only supported for sources in the workspace sources list", *sourceVal)
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
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)

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

func handleVersions(ctx context.Context) error {
	workspacePath := getWorkspacePath(ctx)
	sourceVal := cli.GetOptionalString(ctx, SourceParameter)

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
		if source.Git == nil {
			fmt.Printf("%s LOCAL\n", name)
			continue
		}

		branch := ""
		if source.Git.Branch != nil {
			branch = *source.Git.Branch
		}
		revision := ""
		if source.Git.Revision != nil {
			revision = *source.Git.Revision
		}

		if branch == "" && revision == "" {
			fmt.Printf("%s UNPINNED\n", name)
			continue
		}
		if branch != "" && revision != "" {
			fmt.Printf("%s BRANCH=%s REVISION=%s\n", name, branch, revision)
			continue
		}
		if branch != "" {
			fmt.Printf("%s BRANCH=%s\n", name, branch)
			continue
		}
		fmt.Printf("%s REVISION=%s\n", name, revision)
	}

	return nil
}
