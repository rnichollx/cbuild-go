package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
)

func handleSetCXXVersion(workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup set-cxx-version <version> [<source>]")
	}

	version := args[0]
	var source string
	if len(args) > 1 {
		source = args[1]
	}

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	if source != "" {
		target, ok := ws.Targets[source]
		if !ok {
			return fmt.Errorf("source %s not found in workspace", source)
		}
		target.CxxStandard = &version
		fmt.Printf("Set CXX version for %s to %s\n", source, version)
	} else {
		ws.CXXVersion = version
		fmt.Printf("Set global CXX version to %s\n", version)
	}

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

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
