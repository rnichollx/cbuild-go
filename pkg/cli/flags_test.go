package cli

import (
	"context"
	"testing"
)

func TestParseFlags(t *testing.T) {
	pa := NewParameter("flag-a", ParameterTypeBool, nil, "")
	pb := NewParameter("flag-b", ParameterTypeBool, nil, "")
	pc := NewParameter("flag-c", ParameterTypeString, nil, "")
	pv := NewParameter("verbose-key", ParameterTypeString, nil, "")

	flags := []Flag{
		NewBoolFlag("a", "", pa),
		NewBoolFlag("b", "", pb),
		NewStringFlag("c", "", pc),
		NewStringFlag("", "verbose", pv),
	}

	t.Run("GNU style short args", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-abc", "value"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		valA, _ := GetBool(ctx, pa)
		if valA == nil || !*valA {
			t.Errorf("expected flag-a to be true")
		}
		valB, _ := GetBool(ctx, pb)
		if valB == nil || !*valB {
			t.Errorf("expected flag-b to be true")
		}
		valC, _ := GetString(ctx, pc)
		if valC == nil || *valC != "value" {
			t.Errorf("expected flag-c to be 'value', got %v", valC)
		}
		if len(nonFlagArgs) != 0 {
			t.Errorf("expected no non-flag args, got %v", nonFlagArgs)
		}
	})

	t.Run("Long flags", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"--verbose", "high"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		valV, _ := GetString(ctx, pv)
		if valV == nil || *valV != "high" {
			t.Errorf("expected verbose-key to be 'high', got %v", valV)
		}
		if len(nonFlagArgs) != 0 {
			t.Errorf("expected no non-flag args, got %v", nonFlagArgs)
		}
	})

	t.Run("Non-flag arguments and -- terminator", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-a", "pos1", "--", "-b", "pos2"}
		// Define arguments for opts to allow positional values
		p1 := NewParameter("pos1", ParameterTypeString, nil, "")
		p2 := NewParameter("pos2", ParameterTypeString, nil, "")
		p3 := NewParameter("pos3", ParameterTypeString, nil, "")
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{
			Flags: flags,
			Arguments: []Argument{
				NewStringArgument("p1", p1),
				NewStringArgument("p2", p2),
				NewStringArgument("p3", p3),
			},
		}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		valA, _ := GetBool(ctx, pa)
		if valA == nil || !*valA {
			t.Errorf("expected flag-a to be true")
		}
		valB, _ := GetBool(ctx, pb)
		if valB != nil {
			t.Errorf("expected flag-b to be nil (stopped at --)")
		}

		v1, _ := GetString(ctx, p1)
		if v1 == nil || *v1 != "pos1" {
			t.Errorf("expected pos1 to be 'pos1', got %v", v1)
		}
		v2, _ := GetString(ctx, p2)
		if v2 == nil || *v2 != "-b" {
			t.Errorf("expected pos2 to be '-b', got %v", v2)
		}
		v3, _ := GetString(ctx, p3)
		if v3 == nil || *v3 != "pos2" {
			t.Errorf("expected pos3 to be 'pos2', got %v", v3)
		}

		if len(nonFlagArgs) != 0 {
			t.Errorf("expected 0 non-flag args, got %d", len(nonFlagArgs))
		}
	})

	t.Run("Invalid cluster", func(t *testing.T) {
		args := []string{"-cab", "value"}
		_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flags}, args)
		if err == nil {
			t.Errorf("expected error for value-taking flag in middle of cluster")
		}
	})

	t.Run("Default behavior - double dash removal", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"--verbose", "high", "--", "pos1", "-a"}
		p1 := NewParameter("p1", ParameterTypeString, nil, "")
		p2 := NewParameter("p2", ParameterTypeString, nil, "")
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{
			Flags: flags,
			Arguments: []Argument{
				NewStringArgument("p1", p1),
				NewStringArgument("p2", p2),
			},
		}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		v1, _ := GetString(ctx, p1)
		if v1 == nil || *v1 != "pos1" {
			t.Errorf("expected p1 to be 'pos1', got %v", v1)
		}
		v2, _ := GetString(ctx, p2)
		if v2 == nil || *v2 != "-a" {
			t.Errorf("expected p2 to be '-a', got %v", v2)
		}

		if len(nonFlagArgs) != 0 {
			t.Errorf("expected no non-flag args, got %v", nonFlagArgs)
		}
	})

	t.Run("Duplicate flags", func(t *testing.T) {
		t.Run("Long", func(t *testing.T) {
			args := []string{"--verbose", "val1", "--verbose", "val2"}
			_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flags}, args)
			if err == nil {
				t.Errorf("expected error for duplicate long flag")
			}
		})

		t.Run("Short", func(t *testing.T) {
			args := []string{"-a", "-a"}
			_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flags}, args)
			if err == nil {
				t.Errorf("expected error for duplicate short flag")
			}
		})

		t.Run("Cluster", func(t *testing.T) {
			args := []string{"-aa"}
			_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flags}, args)
			if err == nil {
				t.Errorf("expected error for duplicate short flag in cluster")
			}
		})
	})
}
