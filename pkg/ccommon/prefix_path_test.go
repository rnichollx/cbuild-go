package ccommon

import (
	"strings"
	"testing"
)

func TestCMakePrefixPathEmpty(t *testing.T) {
	w := &Workspace{
		WorkspacePath: "example",
		Targets: map[string]*TargetConfiguration{
			"hello": {
				ProjectType: "executable",
			},
		},
	}

	bp := BuildParameters{
		BuildType: "Debug",
		Toolchain: "system_gcc", // Use a toolchain that exists in the repo
	}

	target := w.Targets["hello"]
	args, err := target.CMakeConfigureArgs(w, "hello", bp)
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	foundPrefixPath := false
	foundModulePath := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-DCMAKE_PREFIX_PATH=") {
			foundPrefixPath = true
			if arg != "-DCMAKE_PREFIX_PATH=" {
				t.Errorf("Expected -DCMAKE_PREFIX_PATH= to be empty, got %s", arg)
			}
		}
		if strings.HasPrefix(arg, "-DCMAKE_MODULE_PATH=") {
			foundModulePath = true
			if arg != "-DCMAKE_MODULE_PATH=" {
				t.Errorf("Expected -DCMAKE_MODULE_PATH= to be empty, got %s", arg)
			}
		}
	}

	if !foundPrefixPath {
		t.Error("Expected -DCMAKE_PREFIX_PATH= to be present")
	}
	if !foundModulePath {
		t.Error("Expected -DCMAKE_MODULE_PATH= to be present")
	}
}
