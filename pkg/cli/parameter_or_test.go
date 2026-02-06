package cli

import (
	"context"
	"reflect"
	"testing"
)

func TestGetStringOr_UsesFallbackWhenUnsetAndNoDefault(t *testing.T) {
	p := Parameter{Key: "name", Type: ParameterTypeString}
	fallback := "fallback"

	v := GetStringOr(context.Background(), p, &fallback)
	if v == nil || *v != "fallback" {
		t.Fatalf("expected fallback value, got %v", v)
	}
}

func TestGetStringOr_DefaultOverridesFallback(t *testing.T) {
	def := "default"
	fallback := "fallback"
	p := Parameter{Key: "name", Type: ParameterTypeString, DefaultValue: &def}

	v := GetStringOr(context.Background(), p, &fallback)
	if v == nil || *v != "default" {
		t.Fatalf("expected default value, got %v", v)
	}
}

func TestGetStringOr_ContextValueOverridesFallback(t *testing.T) {
	p := Parameter{Key: "name", Type: ParameterTypeString}
	fallback := "fallback"
	ctx, err := SetString(context.Background(), p, "from-context")
	if err != nil {
		t.Fatalf("failed to set string: %v", err)
	}

	v := GetStringOr(ctx, p, &fallback)
	if v == nil || *v != "from-context" {
		t.Fatalf("expected context value, got %v", v)
	}
}

func TestGetStringListOr_DefaultOverridesFallback(t *testing.T) {
	def := []string{"default"}
	fallback := []string{"fallback"}
	p := Parameter{Key: "values", Type: ParameterTypeStringList, DefaultValue: &def}

	v := GetStringListOr(context.Background(), p, &fallback)
	if v == nil || !reflect.DeepEqual(*v, def) {
		t.Fatalf("expected default list %v, got %v", def, v)
	}
}

func TestGetString_PanicsWhenUnsetAndNoDefault(t *testing.T) {
	p := Parameter{Key: "name", Type: ParameterTypeString}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when parameter is unset and has no default")
		}
	}()

	_ = GetString(context.Background(), p)
}

func TestGetString_UsesDefaultValue(t *testing.T) {
	def := "default"
	p := Parameter{Key: "name", Type: ParameterTypeString, DefaultValue: &def}

	v := GetString(context.Background(), p)
	if v != "default" {
		t.Fatalf("expected default value, got %v", v)
	}
}
