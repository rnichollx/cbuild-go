package main

import (
	"csetup/pkg/ccommon"
	"os"
)

func main() {

	var ws ccommon.Workspace

	workspace_arg := os.Args[1]

	err := ws.LoadConfig(workspace_arg)
	if err != nil {
		panic(err)
	}
	err = ws.Build()
	if err != nil {
		panic(err)
	}
}
