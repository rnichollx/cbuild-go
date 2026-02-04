package cli

import (
	"context"
	"testing"
)

func TestParsingStyles(t *testing.T) {
	t.Run("ParsingStylePOSIX", func(t *testing.T) {
		// POSIX style: -abc is equivalent to -a -b -c if they are bools,
		// or -a -b -c=... if c takes a value.
		// Wait, the comment says: ParsingStylePOSIX treats single dash as arg+value,
		// e.g. -abc is equivalent to `-a=bc` rather than `-a -b -c`

		pa := NewParameter("a", ParameterTypeBool, nil, "")
		pb := NewParameter("b", ParameterTypeString, nil, "")

		flags := []Flag{
			NewBoolFlag("a", "", pa),
			NewStringFlag("b", "", pb),
		}

		// If ParsingStylePOSIX is used, -abc should be -a with value bc if 'a' took a value.
		// Let's re-read the code for isShortFlag and how it handles clusters.

		// In parsing.go:
		// isShortFlag := func(arg string) bool {
		//   ...
		//   if isLongFlag(arg) || opts.Style == ParsingStyleWindows || opts.Style == ParsingStyleShortWindows || opts.Style == ParsingStyleShort {
		//     return false
		//   }
		//   if arg[:1] == "-" { return true }
		//   return false
		// }
		//
		// If it's short flag, it goes to:
		// else if isShortFlag(arg) {
		//   cluster := arg[1:]
		//   for j := 0; j < len(cluster); j++ {
		//     ...
		//     if requiresValue {
		//       if j != len(cluster)-1 {
		//         valueString := cluster[j+1:]
		//         ...
		//         value = &valueString
		//         j = len(cluster) // consume the rest of the cluster
		//       }
		//     }
		//   }
		// }

		// The comment for ParsingStylePOSIX seems to describe what the code already does for ALL styles that support short flags,
		// UNLESS they are bools.

		t.Run("Short cluster with value", func(t *testing.T) {
			ctx, _, err := ParseFlags(context.Background(), ParseOptions{
				Style: ParsingStylePOSIX,
				Flags: flags,
			}, []string{"-abc"})
			if err == nil {
				// Wait, 'a' is bool, 'b' is string.
				// cluster 'abc':
				// 'a' is found, is bool, so it doesn't take value.
				// 'b' is found, is string, so it takes 'c' as value.
				valA := GetBool(ctx, pa)
				valB := GetString(ctx, pb)
				if valA == nil || !*valA {
					t.Errorf("expected a to be true")
				}
				if valB == nil || *valB != "c" {
					t.Errorf("expected b to be 'c', got %v", valB)
				}
			} else {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	})

	t.Run("ParsingStyleShort", func(t *testing.T) {
		// ParsingStyleShort: long flags are passed like -foo instead of --foo
		pFoo := NewParameter("foo", ParameterTypeString, nil, "")
		flags := []Flag{
			NewStringFlag("", "foo", pFoo),
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Style: ParsingStyleShort,
			Flags: flags,
		}, []string{"-foo", "bar"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val := GetString(ctx, pFoo)
		if val == nil || *val != "bar" {
			t.Errorf("expected foo to be 'bar', got %v", val)
		}
	})

	t.Run("ParsingStyleWindows", func(t *testing.T) {
		// ParsingStyleWindows: /option:value
		pFoo := NewParameter("foo", ParameterTypeString, nil, "")
		flags := []Flag{
			NewStringFlag("", "foo", pFoo),
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Style: ParsingStyleWindows,
			Flags: flags,
		}, []string{"/foo:bar"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val := GetString(ctx, pFoo)
		if val == nil || *val != "bar" {
			t.Errorf("expected foo to be 'bar', got %v", val)
		}
	})

	t.Run("ParsingStyleShortWindows", func(t *testing.T) {
		// ParsingStyleShortWindows: -foo and --% terminator
		pFoo := NewParameter("foo", ParameterTypeString, nil, "")
		pArg := NewParameter("arg", ParameterTypeString, nil, "")
		flags := []Flag{
			NewStringFlag("", "foo", pFoo),
		}
		args := []Argument{
			NewStringArgument("arg", pArg),
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Style:     ParsingStyleShortWindows,
			Flags:     flags,
			Arguments: args,
		}, []string{"-foo", "bar", "--%", "-baz"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val := GetString(ctx, pFoo)
		if val == nil || *val != "bar" {
			t.Errorf("expected foo to be 'bar', got %v", val)
		}

		argVal := GetString(ctx, pArg)
		if argVal == nil || *argVal != "-baz" {
			t.Errorf("expected arg to be '-baz', got %v", argVal)
		}
	})
}
