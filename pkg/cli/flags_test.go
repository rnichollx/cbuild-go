package cli

import (
	"context"
	"testing"
)

func TestParseFlags(t *testing.T) {
	flags := []Flag{
		&BoolFlag{short: "a", key: "flag-a"},
		&BoolFlag{short: "b", key: "flag-b"},
		&StringFlag{short: "c", key: "flag-c"},
		&StringFlag{long: "verbose", key: "verbose-key"},
	}

	t.Run("GNU style short args", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-abc", "value"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.Value(FlagKey("flag-a")) != "true" {
			t.Errorf("expected flag-a to be true")
		}
		if ctx.Value(FlagKey("flag-b")) != "true" {
			t.Errorf("expected flag-b to be true")
		}
		if ctx.Value(FlagKey("flag-c")) != "value" {
			t.Errorf("expected flag-c to be 'value', got %v", ctx.Value(FlagKey("flag-c")))
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

		if ctx.Value(FlagKey("verbose-key")) != "high" {
			t.Errorf("expected verbose-key to be 'high', got %v", ctx.Value(FlagKey("verbose-key")))
		}
		if len(nonFlagArgs) != 0 {
			t.Errorf("expected no non-flag args, got %v", nonFlagArgs)
		}
	})

	t.Run("Non-flag arguments and -- terminator", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-a", "pos1", "--", "-b", "pos2"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.Value(FlagKey("flag-a")) != "true" {
			t.Errorf("expected flag-a to be true")
		}
		if ctx.Value(FlagKey("flag-b")) != nil {
			t.Errorf("expected flag-b to be nil (stopped at --)")
		}

		expectedNonFlagArgs := []string{"pos1", "-b", "pos2"}
		if len(nonFlagArgs) != len(expectedNonFlagArgs) {
			t.Fatalf("expected %d non-flag args, got %d", len(expectedNonFlagArgs), len(nonFlagArgs))
		}
		for i, v := range expectedNonFlagArgs {
			if nonFlagArgs[i] != v {
				t.Errorf("expected non-flag arg %d to be %s, got %s", i, v, nonFlagArgs[i])
			}
		}
	})

	t.Run("Invalid cluster", func(t *testing.T) {
		args := []string{"-cab", "value"}
		_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flags}, args)
		if err == nil {
			t.Errorf("expected error for value-taking flag in middle of cluster")
		}
	})

	t.Run("Allow unknown flags - long", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"--verbose", "high", "--unknown", "arg1"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags, AllowUnknownFlags: true}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.Value(FlagKey("verbose-key")) != "high" {
			t.Errorf("expected verbose-key to be 'high', got %v", ctx.Value(FlagKey("verbose-key")))
		}

		expectedNonFlagArgs := []string{"--unknown", "arg1"}
		if len(nonFlagArgs) != len(expectedNonFlagArgs) {
			t.Fatalf("expected %d non-flag args, got %v", len(expectedNonFlagArgs), nonFlagArgs)
		}
		for i, v := range expectedNonFlagArgs {
			if nonFlagArgs[i] != v {
				t.Errorf("expected arg %d to be %s, got %s", i, v, nonFlagArgs[i])
			}
		}
	})

	t.Run("Allow unknown flags - short cluster", func(t *testing.T) {
		ctx := context.Background()
		// -a and -b are known, -x and -y are unknown
		args := []string{"-axby", "pos1"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags, AllowUnknownFlags: true}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.Value(FlagKey("flag-a")) != "true" {
			t.Errorf("expected flag-a to be true")
		}
		if ctx.Value(FlagKey("flag-b")) != "true" {
			t.Errorf("expected flag-b to be true")
		}

		expectedNonFlagArgs := []string{"-x", "-y", "pos1"}
		if len(nonFlagArgs) != len(expectedNonFlagArgs) {
			t.Fatalf("expected %d non-flag args, got %v", len(expectedNonFlagArgs), nonFlagArgs)
		}
		for i, v := range expectedNonFlagArgs {
			if nonFlagArgs[i] != v {
				t.Errorf("expected arg %d to be %s, got %s", i, v, nonFlagArgs[i])
			}
		}
	})

	t.Run("Allow unknown flags - short cluster with value", func(t *testing.T) {
		ctx := context.Background()
		// -a and -c are known, -x is unknown. -c takes value.
		args := []string{"-axc", "val", "pos1"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags, AllowUnknownFlags: true}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.Value(FlagKey("flag-a")) != "true" {
			t.Errorf("expected flag-a to be true")
		}
		if ctx.Value(FlagKey("flag-c")) != "val" {
			t.Errorf("expected flag-c to be 'val', got %v", ctx.Value(FlagKey("flag-c")))
		}

		expectedNonFlagArgs := []string{"-x", "pos1"}
		if len(nonFlagArgs) != len(expectedNonFlagArgs) {
			t.Fatalf("expected %d non-flag args, got %v", len(expectedNonFlagArgs), nonFlagArgs)
		}
		for i, v := range expectedNonFlagArgs {
			if nonFlagArgs[i] != v {
				t.Errorf("expected arg %d to be %s, got %s", i, v, nonFlagArgs[i])
			}
		}
	})

	t.Run("Allow unknown flags - double dash preservation", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"--verbose", "high", "--", "pos1", "-a"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags, AllowUnknownFlags: true}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx.Value(FlagKey("verbose-key")) != "high" {
			t.Errorf("expected verbose-key to be 'high', got %v", ctx.Value(FlagKey("verbose-key")))
		}

		expectedNonFlagArgs := []string{"--", "pos1", "-a"}
		if len(nonFlagArgs) != len(expectedNonFlagArgs) {
			t.Fatalf("expected %d non-flag args, got %v", len(expectedNonFlagArgs), nonFlagArgs)
		}
		for i, v := range expectedNonFlagArgs {
			if nonFlagArgs[i] != v {
				t.Errorf("expected arg %d to be %s, got %s", i, v, nonFlagArgs[i])
			}
		}
	})

	t.Run("Default behavior - double dash removal", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"--verbose", "high", "--", "pos1", "-a"}
		ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedNonFlagArgs := []string{"pos1", "-a"}
		if len(nonFlagArgs) != len(expectedNonFlagArgs) {
			t.Fatalf("expected %d non-flag args, got %v", len(expectedNonFlagArgs), nonFlagArgs)
		}
		for i, v := range expectedNonFlagArgs {
			if nonFlagArgs[i] != v {
				t.Errorf("expected arg %d to be %s, got %s", i, v, nonFlagArgs[i])
			}
		}
	})
}
