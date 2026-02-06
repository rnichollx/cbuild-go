package cli

import (
	"context"
	"reflect"
	"testing"
)

func TestParsingLists(t *testing.T) {
	t.Run("RPNX style list specifier", func(t *testing.T) {
		pList := Parameter{Key: "list", Type: ParameterTypeStringList, DefaultValue: nil, Description: ""}
		flags := []Flag{
			Flag{Short: "l", Long: "list", Parameter: pList},
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Style: ParsingStyleRPNX,
			Flags: flags,
		}, []string{"--list", "[", "+val1", "+val2", "]"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val := GetOptionalStringList(ctx, pList)
		expected := []string{"val1", "val2"}
		if val == nil || !reflect.DeepEqual(*val, expected) {
			t.Errorf("expected list to be %v, got %v", expected, val)
		}
	})

	t.Run("Greedy list flag", func(t *testing.T) {
		pList := Parameter{Key: "list", Type: ParameterTypeStringList, DefaultValue: nil, Description: ""}
		flags := []Flag{
			Flag{
				Short:     "l",
				Long:      "list",
				Parameter: pList,
				Greedy:    true,
			},
		}

		ctx := context.Background()
		_, err := ParseFlagsAndArgs(ParseOptions{
			Flags: flags,
		}, ParseInput{Ctx: ctx, Tokens: []string{"--list", "val1", "val2", "--other"}})
		// wait, --other is unknown, so it should error
		if err == nil {
			t.Errorf("expected error for unknown flag --other")
		}

		// Try with valid args after greedy
		pOther := Parameter{Key: "other", Type: ParameterTypeBool, DefaultValue: nil, Description: ""}
		flags = append(flags, Flag{Short: "o", Long: "other", Parameter: pOther})

		ctx, _, err = ParseFlags(ctx, ParseOptions{
			Flags: flags,
		}, []string{"--list", "val1", "val2", "--other"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val := GetOptionalStringList(ctx, pList)
		expected := []string{"val1", "val2"}
		if val == nil || !reflect.DeepEqual(*val, expected) {
			t.Errorf("expected list to be %v, got %v", expected, val)
		}

		otherVal := GetOptionalBool(ctx, pOther)
		if otherVal == nil || !*otherVal {
			t.Errorf("expected other to be true")
		}
	})

	t.Run("Argument list with separator", func(t *testing.T) {
		pList := Parameter{Key: "list", Type: ParameterTypeStringList, DefaultValue: nil, Description: ""}
		sep := ","
		args := []Argument{
			Argument{
				Name:      "list",
				Parameter: pList,
				Separator: &sep,
			},
		}

		ctx, _, err := ParseFlags(context.Background(), ParseOptions{
			Arguments: args,
		}, []string{"val1,val2,val3"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val := GetOptionalStringList(ctx, pList)
		expected := []string{"val1", "val2", "val3"}
		if val == nil || !reflect.DeepEqual(*val, expected) {
			t.Errorf("expected list to be %v, got %v", expected, val)
		}
	})
}
