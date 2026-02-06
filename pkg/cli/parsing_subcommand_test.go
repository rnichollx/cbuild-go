package cli

import (
	"context"
	"reflect"
	"testing"
)

func TestSubcommands(t *testing.T) {
	pSub := Parameter{Key: "sub", Type: ParameterTypeBool, DefaultValue: nil, Description: "sub flag"}
	pArg := Parameter{Key: "arg", Type: ParameterTypeString, DefaultValue: nil, Description: "arg 1"}

	opts := ParseOptions{
		Subcommands: map[string]SubcommandParseOptions{
			"cmd": {
				ParseOptions: &ParseOptions{
					Flags: []Flag{
						Flag{Short: "s", Long: "sub", Parameter: pSub},
					},
					Arguments: []Argument{
						Argument{Name: "arg", Parameter: pArg},
					},
				},
			},
		},
	}

	input := ParseInput{
		Ctx:    context.Background(),
		Tokens: []string{"cmd", "--sub", "foo"},
	}

	result, err := ParseFlagsAndArgs(opts, input)
	if err != nil {
		t.Fatalf("ParseFlagsAndArgs failed: %v", err)
	}

	if !reflect.DeepEqual(result.Subcommands, []string{"cmd"}) {
		t.Errorf("Expected subcommands [cmd], got %v", result.Subcommands)
	}

	subVal := GetOptionalBool(result.Ctx, pSub)
	if subVal == nil || !*subVal {
		t.Errorf("Expected sub flag to be true")
	}

	argVal := GetOptionalString(result.Ctx, pArg)
	if argVal == nil || *argVal != "foo" {
		t.Errorf("Expected arg to be 'foo', got %v", argVal)
	}
}

func TestNestedSubcommands(t *testing.T) {
	p1 := Parameter{Key: "p1", Type: ParameterTypeBool, DefaultValue: nil, Description: ""}
	p2 := Parameter{Key: "p2", Type: ParameterTypeBool, DefaultValue: nil, Description: ""}

	opts := ParseOptions{
		Subcommands: map[string]SubcommandParseOptions{
			"sub1": {
				ParseOptions: &ParseOptions{
					Subcommands: map[string]SubcommandParseOptions{
						"sub2": {
							ParseOptions: &ParseOptions{
								Flags: []Flag{
									Flag{Short: "2", Long: "p2", Parameter: p2},
								},
							},
						},
					},
					Flags: []Flag{
						Flag{Short: "1", Long: "p1", Parameter: p1},
					},
				},
			},
		},
	}

	input := ParseInput{
		Ctx:    context.Background(),
		Tokens: []string{"sub1", "--p1", "sub2", "--p2"},
	}

	result, err := ParseFlagsAndArgs(opts, input)
	if err != nil {
		t.Fatalf("ParseFlagsAndArgs failed: %v", err)
	}

	if !reflect.DeepEqual(result.Subcommands, []string{"sub1", "sub2"}) {
		t.Errorf("Expected subcommands [sub1 sub2], got %v", result.Subcommands)
	}

	v1 := GetOptionalBool(result.Ctx, p1)
	if v1 == nil || !*v1 {
		t.Errorf("Expected p1 to be true")
	}

	v2 := GetOptionalBool(result.Ctx, p2)
	if v2 == nil || !*v2 {
		t.Errorf("Expected p2 to be true")
	}
}

func TestStopParsing(t *testing.T) {
	opts := ParseOptions{
		Subcommands: map[string]SubcommandParseOptions{
			"run": {
				StopParsing: true,
			},
		},
	}

	input := ParseInput{
		Ctx:    context.Background(),
		Tokens: []string{"run", "--any", "args", "here"},
	}

	result, err := ParseFlagsAndArgs(opts, input)
	if err != nil {
		t.Fatalf("ParseFlagsAndArgs failed: %v", err)
	}

	if !reflect.DeepEqual(result.Subcommands, []string{"run"}) {
		t.Errorf("Expected subcommands [run], got %v", result.Subcommands)
	}

	expectedUnparsed := []string{"--any", "args", "here"}
	if !reflect.DeepEqual(result.Unparsed, expectedUnparsed) {
		t.Errorf("Expected unparsed %v, got %v", expectedUnparsed, result.Unparsed)
	}
}

func TestSubcommandRestriction(t *testing.T) {
	pArg1 := Parameter{Key: "arg1", Type: ParameterTypeString, DefaultValue: nil, Description: ""}
	pArg2 := Parameter{Key: "arg2", Type: ParameterTypeString, DefaultValue: nil, Description: ""}
	opts := ParseOptions{
		Arguments: []Argument{
			Argument{Name: "arg1", Parameter: pArg1},
			Argument{Name: "arg2", Parameter: pArg2},
		},
		Subcommands: map[string]SubcommandParseOptions{
			"cmd": {
				StopParsing: true,
			},
		},
	}

	// cmd is NOT a subcommand here because it's after a positional argument
	input := ParseInput{
		Ctx:    context.Background(),
		Tokens: []string{"foo", "cmd"},
	}

	result, err := ParseFlagsAndArgs(opts, input)
	if err != nil {
		t.Fatalf("ParseFlagsAndArgs failed: %v", err)
	}

	if len(result.Subcommands) != 0 {
		t.Errorf("Expected no subcommands, got %v", result.Subcommands)
	}

	arg1Val := GetOptionalString(result.Ctx, pArg1)
	if arg1Val == nil || *arg1Val != "foo" {
		t.Errorf("Expected first arg to be 'foo'")
	}

	arg2Val := GetOptionalString(result.Ctx, pArg2)
	if arg2Val == nil || *arg2Val != "cmd" {
		t.Errorf("Expected second arg to be 'cmd', got %v", arg2Val)
	}
}
