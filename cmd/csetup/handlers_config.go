package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
	"os"
)

func handleSetCXXVersion(workspacePath string, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: csetup set-cxx-version <source> <version>")
		os.Exit(1)
	}

	source := args[0]
	version := args[1]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	ws.CXXVersion = version
	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Set CXX version to %s (source argument %s was ignored as it is currently global)\n", version, source)
}

func handleEnableStaging(workspacePath string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: csetup enable-staging <source>")
		os.Exit(1)
	}

	source := args[0]

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

	staged := true
	target.Staged = &staged

	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Enabled staging for %s\n", source)
}

func handleDisableStaging(workspacePath string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: csetup disable-staging <source>")
		os.Exit(1)
	}

	source := args[0]

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

	staged := false
	target.Staged = &staged

	err = ws.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Disabled staging for %s\n", source)
}
