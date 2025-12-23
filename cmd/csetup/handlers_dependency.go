package main

import (
	"context"
	"fmt"
	"gitlab.com/rpnx/cbuild-go/pkg/ccommon"
)

func handleAddDependency(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: csetup add-dependency <source> <depname>")
	}

	source := args[0]
	depname := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	target, ok := ws.Targets[source]
	if !ok {
		return fmt.Errorf("source %s not found in workspace", source)
	}

	// Check if dependency already exists
	for _, d := range target.Depends {
		if d == depname {
			fmt.Printf("Dependency %s already exists for %s\n", depname, source)
			return nil
		}
	}

	target.Depends = append(target.Depends, depname)
	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Added dependency %s to %s\n", depname, source)
	return nil
}

func handleRemoveDependency(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: csetup remove-dependency <source> <depname>")
	}

	source := args[0]
	depname := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	target, ok := ws.Targets[source]
	if !ok {
		return fmt.Errorf("source %s not found in workspace", source)
	}

	newDepends := []string{}
	found := false
	for _, d := range target.Depends {
		if d == depname {
			found = true
			continue
		}
		newDepends = append(newDepends, d)
	}

	if !found {
		fmt.Printf("Dependency %s not found for %s\n", depname, source)
		return nil
	}

	target.Depends = newDepends
	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed dependency %s from %s\n", depname, source)
	return nil
}
