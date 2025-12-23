package cli

import "context"

type Subcommand struct {
	Description       string
	HelpText          string
	AcceptsFlags      []Flag
	AllowArgs         bool
	AllowUnknownFlags bool
	Exec              func(ctx context.Context, args []string) error
}
