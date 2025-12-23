package main

import (
	"cbuild-go/pkg/ccommon"
	"context"
	"fmt"
)

func handleSetCXXVersion(ctx context.Context, workspacePath string, args []string) error {
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

func handleEnableStaging(ctx context.Context, workspacePath string, args []string) error {
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

func handleDisableStaging(ctx context.Context, workspacePath string, args []string) error {
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

func handleAddConfig(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup add-config <configname>")
	}

	configName := args[0]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	for _, cfg := range ws.Configurations {
		if cfg == configName {
			fmt.Printf("Configuration %s already exists\n", configName)
			return nil
		}
	}

	ws.Configurations = append(ws.Configurations, configName)

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Added configuration %s\n", configName)
	return nil
}

func handleRemoveConfig(ctx context.Context, workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup remove-config <configname>")
	}

	configName := args[0]

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	found := false
	newConfigs := []string{}
	for _, cfg := range ws.Configurations {
		if cfg == configName {
			found = true
			continue
		}
		newConfigs = append(newConfigs, cfg)
	}

	if !found {
		return fmt.Errorf("configuration %s not found", configName)
	}

	ws.Configurations = newConfigs

	err = ws.Save()
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed configuration %s\n", configName)
	return nil
}
