package cli

import (
	"context"
	"testing"
)

func TestShortFlagValues(t *testing.T) {
	pd := NewParameter("D", ParameterTypeString, nil, "")
	flags := []Flag{
		NewStringFlag("D", "", pd),
	}

	t.Run("-Dfoo style", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-Dfoo"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foo" {
			t.Errorf("expected D to be 'foo', got %v", val)
		}
	})

	t.Run("-D=foobar style", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-D=foobar"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foobar" {
			t.Errorf("expected D to be 'foobar', got %v", val)
		}
	})

	t.Run("-D foobar style (standard)", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-D", "foobar"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foobar" {
			t.Errorf("expected D to be 'foobar', got %v", val)
		}
	})

	t.Run("Windows style /D:foobar", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"/D:foobar"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags, Style: ParsingStyleWindows}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foobar" {
			t.Errorf("expected D to be 'foobar', got %v", val)
		}
	})

	t.Run("Windows style /Dfoobar", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"/Dfoobar"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags, Style: ParsingStyleWindows}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foobar" {
			t.Errorf("expected D to be 'foobar', got %v", val)
		}
	})

	t.Run("POSIX style -D=foobar (no cluster)", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-D=foobar"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags, Style: ParsingStylePOSIX}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foobar" {
			t.Errorf("expected D to be 'foobar', got %v", val)
		}
	})

	t.Run("POSIX style -Dfoobar (no cluster)", func(t *testing.T) {
		ctx := context.Background()
		args := []string{"-Dfoobar"}
		ctx, _, err := ParseFlags(ctx, ParseOptions{Flags: flags, Style: ParsingStylePOSIX}, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetString(ctx, pd)
		if val == nil || *val != "foobar" {
			t.Errorf("expected D to be 'foobar', got %v", val)
		}
	})
}
