package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
)

func handleSetCXXVersion(workspacePath string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: csetup set-cxx-version <source> <version>")
	}

	source := args[0]
	version := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	ws.CXXVersion = version
	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Set CXX version to %s (source argument %s was ignored as it is currently global)\n", version, source)
	return nil
}

func handleEnableStaging(workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup enable-staging <source>")
	}

	source := args[0]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	target, ok := ws.Targets[source]
	if !ok {
		return fmt.Errorf("source %s not found in workspace", source)
	}

	staged := true
	target.Staged = &staged

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Enabled staging for %s\n", source)
	return nil
}

func handleDisableStaging(workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup disable-staging <source>")
	}

	source := args[0]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	target, ok := ws.Targets[source]
	if !ok {
		return fmt.Errorf("source %s not found in workspace", source)
	}

	staged := false
	target.Staged = &staged

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Disabled staging for %s\n", source)
	return nil
}
