package cli

import (
	"context"
	"fmt"
	"testing"
)

func TestRunner(t *testing.T) {
	globalFlag := NewStringFlag("w", "workspace", "workspace-key", "global workspace")
	subFlag := NewBoolFlag("v", "verbose", "verbose-key", "subcommand verbose")

	// Create a required version of workspace flag for a specific subcommand
	reqWorkspaceFlag := NewRequiredStringFlag("w", "workspace", "workspace-key", "required workspace for sub")

	runner := &Runner{
		Name:        "testapp",
		GlobalFlags: []Flag{globalFlag, NewBoolFlag("h", "help", "help", "show help")},
		Subcommands: map[string]*Subcommand{
			"sub": {
				Name:         "sub",
				Description:  "a subcommand",
				AcceptsFlags: []Flag{subFlag, reqWorkspaceFlag},
				Exec: func(ctx context.Context, args []string) error {
					ws := GetString(ctx, "workspace-key")
					if ws == "" {
						return fmt.Errorf("workspace not set")
					}
					return nil
				},
			},
		},
	}

	t.Run("Global flag after subcommand", func(t *testing.T) {
		args := []string{"sub", "-w", "./ws"}
		err := runner.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Global flag before subcommand", func(t *testing.T) {
		args := []string{"-w", "./ws", "sub"}
		err := runner.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Required flag for subcommand", func(t *testing.T) {
		args := []string{"sub"}
		err := runner.Run(context.Background(), args)
		if err == nil {
			t.Errorf("expected error for missing required flag in subcommand")
		}
	})

	t.Run("Default subcommand", func(t *testing.T) {
		runnerWithDefault := &Runner{
			Name:          "testapp",
			GlobalFlags:   []Flag{globalFlag},
			Subcommands:   runner.Subcommands,
			DefaultSubcmd: "sub",
		}
		args := []string{"-w", "./ws"}
		err := runnerWithDefault.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error with default subcommand: %v", err)
		}
	})

	t.Run("Subcommand help", func(t *testing.T) {
		// This is hard to test automatically since it prints to stdout,
		// but we can at least ensure it doesn't return an error.
		args := []string{"-h", "sub"}
		err := runner.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		args = []string{"sub", "-h"}
		err = runner.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		args = []string{"-w", "./ws", "sub", "-h"}
		err = runner.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Help with default subcommand", func(t *testing.T) {
		runnerWithDefault := &Runner{
			Name:          "testapp",
			GlobalFlags:   []Flag{NewBoolFlag("h", "help", "help", "show help")},
			Subcommands:   runner.Subcommands,
			DefaultSubcmd: "sub",
		}

		// When ONLY -h is provided, it should show general help (list subcommands)
		// even if there is a default subcommand.
		// Currently this is hard to verify via error code because Run returns nil for help.
		// But we can check if it tries to run the default subcommand (it shouldn't).

		args := []string{"-h"}
		err := runnerWithDefault.Run(context.Background(), args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
