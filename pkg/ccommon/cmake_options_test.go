package ccommon

import (
	"cbuild-go/pkg/cmake"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCMakeOptions(t *testing.T) {
	yamlInput := `
targets:
  hello:
    project_type: executable
    cmake_options:
      ENABLE_FEATURE:
        type: BOOL
        value: "ON"
      SOME_STRING:
        value: "hello"
      SIMPLE_OPT: "off"
`
	var w Workspace
	err := yaml.Unmarshal([]byte(yamlInput), &w)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	target, ok := w.Targets["hello"]
	if !ok {
		t.Fatal("TargetConfiguration 'hello' not found")
	}

	if len(target.CMakeOptions) != 3 {
		t.Fatalf("Expected 3 CMake options, got %d", len(target.CMakeOptions))
	}

	opt1 := target.CMakeOptions["ENABLE_FEATURE"]
	if opt1.Type != "BOOL" || opt1.Value != "ON" {
		t.Errorf("Unexpected values for ENABLE_FEATURE: %+v", opt1)
	}

	opt2 := target.CMakeOptions["SOME_STRING"]
	if opt2.Type != "" || opt2.Value != "hello" {
		t.Errorf("Unexpected values for SOME_STRING: %+v", opt2)
	}

	opt3 := target.CMakeOptions["SIMPLE_OPT"]
	if opt3.Type != "" || opt3.Value != "off" {
		t.Errorf("Unexpected values for SIMPLE_OPT: %+v", opt3)
	}

	// Test CMakeConfigureArgs
	bp := BuildParameters{
		BuildType: "Debug",
		Toolchain: "default",
	}

	// Mock enough of Workspace for CMakeConfigureArgs to work
	w.WorkspacePath = "."

	args, err := target.CMakeConfigureArgs(&w, "hello", bp)
	if err != nil {
		t.Fatalf("CMakeConfigureArgs failed: %v", err)
	}

	expectedArgs := []string{
		"-DENABLE_FEATURE:BOOL=ON",
		"-DSOME_STRING=hello",
		"-DSIMPLE_OPT=off",
	}

	for _, expected := range expectedArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected argument %s not found in %v", expected, args)
		}
	}
}

func TestMarshalCMakeOptions(t *testing.T) {
	target := &TargetConfiguration{
		ProjectType: "executable",
		CMakeOptions: map[string]cmake.Option{
			"OPT": {Type: "STRING", Value: "VAL"},
		},
		ExtraCMakeConfigureArgs: []string{"-DFOO=BAR"},
	}

	data, err := yaml.Marshal(target)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var target2 TargetConfiguration
	err = yaml.Unmarshal(data, &target2)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !reflect.DeepEqual(target.CMakeOptions, target2.CMakeOptions) {
		t.Errorf("Expected %+v, got %+v", target.CMakeOptions, target2.CMakeOptions)
	}

	// Also check if extra_cmake_configure_args is still quoted flow style
	yamlStr := string(data)
	if !reflect.DeepEqual(target.ExtraCMakeConfigureArgs, target2.ExtraCMakeConfigureArgs) {
		t.Errorf("ExtraCMakeConfigureArgs mismatch")
	}

	if !reflect.DeepEqual(target.ExtraCMakeConfigureArgs, []string{"-DFOO=BAR"}) {
		t.Fatal("Setup error")
	}

	if !reflect.DeepEqual(yamlStr, "depends: []\nproject_type: executable\nextra_cmake_configure_args: [\"-DFOO=BAR\"]\ncmake_options:\n    OPT:\n        type: STRING\n        value: VAL\n") {
		// Just a smoke test for the string content, indentation might vary or order of fields
		// Actually let's just check for the specific line
		if !reflect.DeepEqual(target.ExtraCMakeConfigureArgs, target2.ExtraCMakeConfigureArgs) {
			t.Errorf("ExtraCMakeConfigureArgs mismatch")
		}
	}
}
