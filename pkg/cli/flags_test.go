package cli

import (
	"context"
	"testing"
)

func TestParseFlags(t *testing.T) {
	flags := []Flag{
		NewBoolFlag("a", "", "flag-a", ""),
		NewBoolFlag("b", "", "flag-b", ""),
		NewStringFlag("c", "", "flag-c", ""),
		NewStringFlag("", "verbose", "verbose-key", ""),
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

	t.Run("FromArgument", func(t *testing.T) {
		faFlag := NewStringFlagFromArgument("f", "from", "fa-key", "description")
		flagsWithFA := append(flags, faFlag)

		t.Run("Implicit", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"value", "-a"}
			ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flagsWithFA}, args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if GetString(ctx, "fa-key") != "value" {
				t.Errorf("expected fa-key to be 'value', got %v", GetString(ctx, "fa-key"))
			}
			if GetBool(ctx, "flag-a") != true {
				t.Errorf("expected flag-a to be true")
			}
			if len(nonFlagArgs) != 0 {
				t.Errorf("expected 0 non-flag args, got %v", nonFlagArgs)
			}
		})

		t.Run("Explicit", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"-f", "explicit", "implicit"}
			ctx, nonFlagArgs, err := ParseFlags(ctx, ParseOptions{Flags: flagsWithFA}, args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if GetString(ctx, "fa-key") != "explicit" {
				t.Errorf("expected fa-key to be 'explicit', got %v", GetString(ctx, "fa-key"))
			}
			if len(nonFlagArgs) != 1 || nonFlagArgs[0] != "implicit" {
				t.Errorf("expected 1 non-flag arg 'implicit', got %v", nonFlagArgs)
			}
		})

		t.Run("Multiple FromArgument flags error", func(t *testing.T) {
			faFlag2 := NewStringFlagFromArgument("g", "from2", "fa-key2", "description")
			flagsWithTwoFA := append(flagsWithFA, faFlag2)
			_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flagsWithTwoFA}, []string{})
			if err == nil {
				t.Errorf("expected error for multiple FromArgument flags")
			}
		})
	})

	t.Run("Required flags", func(t *testing.T) {
		reqFlag := NewRequiredStringFlag("r", "required", "req-key", "description")
		flagsWithReq := append(flags, reqFlag)

		t.Run("Missing required flag", func(t *testing.T) {
			_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flagsWithReq}, []string{})
			if err == nil {
				t.Errorf("expected error for missing required flag")
			}
		})

		t.Run("Provided required flag", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"-r", "val"}
			ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flagsWithReq}, args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if GetString(ctx, "req-key") != "val" {
				t.Errorf("expected req-key to be 'val', got %v", GetString(ctx, "req-key"))
			}
		})

		t.Run("Required BoolFlag", func(t *testing.T) {
			reqBool := NewRequiredBoolFlag("R", "req-bool", "req-bool-key", "")
			flagsWithReqBool := append(flags, reqBool)

			t.Run("Missing", func(t *testing.T) {
				_, _, err := ParseFlags(context.Background(), ParseOptions{Flags: flagsWithReqBool}, []string{})
				if err == nil {
					t.Errorf("expected error for missing required bool flag")
				}
			})

			t.Run("Provided", func(t *testing.T) {
				ctx := context.Background()
				args := []string{"-R"}
				ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flagsWithReqBool}, args)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if GetBool(ctx, "req-bool-key") != true {
					t.Errorf("expected req-bool-key to be true")
				}
			})
		})
	})
}
