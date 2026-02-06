package csetupapp

import (
	"context"
	"fmt"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
)

func handleHelp(ctx context.Context) error {
	subcmdName := ""
	subcmdNameRaw := cli.GetOptionalString(ctx, HelpSubcommandParameter)
	if subcmdNameRaw != nil {
		subcmdName = *subcmdNameRaw
	}

	if subcmdName != "" {
		if _, ok := CSetup.Subcommands[subcmdName]; !ok {
			return fmt.Errorf("unknown subcommand %q", subcmdName)
		}
	}

	CSetup.PrintUsage(subcmdName)
	return nil
}
