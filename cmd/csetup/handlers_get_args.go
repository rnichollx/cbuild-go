package main

import (
	"cbuild-go/pkg/ccommon"
	"fmt"
	"strings"
)

func handleGetArgs(workspacePath string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: csetup get-args <target> [-T|--toolchain <toolchain>] [-c|--config <type>]")
	}

	targetName := args[0]
	toolchain := "default"
	buildType := "Debug"

	for i := 1; i < len(args); i++ {
		if (args[i] == "-T" || args[i] == "--toolchain") && i+1 < len(args) {
			toolchain = args[i+1]
			i++
		} else if (args[i] == "-c" || args[i] == "--config" || args[i] == "--build-type") && i+1 < len(args) {
			buildType = args[i+1]
			i++
		}
	}

	ws := &ccommon.Workspace{}
	err := ws.Load(workspacePath)
	if err != nil {
		return fmt.Errorf("error loading workspace: %w", err)
	}

	target, ok := ws.Targets[targetName]
	if !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	bp := ccommon.BuildParameters{
		Toolchain: toolchain,
		BuildType: buildType,
	}

	fullArgs, err := target.CMakeConfigureArgs(ws, targetName, bp)
	if err != nil {
		return fmt.Errorf("error getting cmake args: %w", err)
	}

	filteredArgs := []string{}
	for i := 0; i < len(fullArgs); i++ {
		arg := fullArgs[i]
		if arg == "-S" || arg == "-B" {
			i++ // skip the next argument too (the path)
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	fmt.Println(strings.Join(filteredArgs, " "))
	return nil
}
