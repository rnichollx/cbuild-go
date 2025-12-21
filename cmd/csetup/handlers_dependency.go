package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
	"os"
)

func handleAddDependency(workspacePath string, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: csetup add-dependency <source> <depname>")
		os.Exit(1)
	}

	source := args[0]
	depname := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	target, ok := ws.Targets[source]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
	}

	// Check if dependency already exists
	for _, d := range target.Depends {
		if d == depname {
			fmt.Printf("Dependency %s already exists for %s\n", depname, source)
			return
		}
	}

	target.Depends = append(target.Depends, depname)
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added dependency %s to %s\n", depname, source)
}

func handleRemoveDependency(workspacePath string, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: csetup remove-dependency <source> <depname>")
		os.Exit(1)
	}

	source := args[0]
	depname := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	target, ok := ws.Targets[source]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: source %s not found in workspace\n", source)
		os.Exit(1)
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
		return
	}

	target.Depends = newDepends
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed dependency %s from %s\n", depname, source)
}
