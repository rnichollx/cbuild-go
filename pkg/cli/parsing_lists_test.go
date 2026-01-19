package cli

import (
	"context"
	"reflect"
	"testing"
)

func TestParsingLists(t *testing.T) {
	t.Run("RPNX style list specifier", func(t *testing.T) {
		pList := NewParameter("list", ParameterTypeStringList, nil, "", false)
		flags := []Flag{
			NewStringFlag("l", "list", pList),
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Style: ParsingStyleRPNX,
			Flags: flags,
		}, []string{"--list", "[", "+val1", "+val2", "]"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetStringList(ctx, pList)
		expected := []string{"val1", "val2"}
		if val == nil || !reflect.DeepEqual(*val, expected) {
			t.Errorf("expected list to be %v, got %v", expected, val)
		}
	})

	t.Run("Greedy list flag", func(t *testing.T) {
		pList := NewParameter("list", ParameterTypeStringList, nil, "", false)
		flags := []Flag{
			&baseFlag{
				short:     "l",
				long:      "list",
				parameter: pList,
				greedy:    true,
			},
		}

		_, err := ParseFlagsAndArgs(ParseOptions{
			Flags: flags,
		}, ParseInput{Tokens: []string{"--list", "val1", "val2", "--other"}})
		// wait, --other is unknown, so it should error
		if err == nil {
			t.Errorf("expected error for unknown flag --other")
		}

		// Try with valid args after greedy
		pOther := NewParameter("other", ParameterTypeBool, nil, "", false)
		flags = append(flags, NewBoolFlag("o", "other", pOther))

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Flags: flags,
		}, []string{"--list", "val1", "val2", "--other"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetStringList(ctx, pList)
		expected := []string{"val1", "val2"}
		if val == nil || !reflect.DeepEqual(*val, expected) {
			t.Errorf("expected list to be %v, got %v", expected, val)
		}

		otherVal, _ := GetBool(ctx, pOther)
		if otherVal == nil || !*otherVal {
			t.Errorf("expected other to be true")
		}
	})

	t.Run("Argument list with separator", func(t *testing.T) {
		pList := NewParameter("list", ParameterTypeStringList, nil, "", false)
		sep := ","
		args := []Argument{
			&baseArgument{
				name:      "list",
				parameter: pList,
				separator: &sep,
			},
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Arguments: args,
		}, []string{"val1,val2,val3"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := GetStringList(ctx, pList)
		expected := []string{"val1", "val2", "val3"}
		if val == nil || !reflect.DeepEqual(*val, expected) {
			t.Errorf("expected list to be %v, got %v", expected, val)
		}
	})
}
