package cbuildapp

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
		if _, ok := CBuild.Subcommands[subcmdName]; !ok {
			return fmt.Errorf("unknown subcommand %q", subcmdName)
		}
	}

	CBuild.PrintUsage(subcmdName)
	return nil
}
