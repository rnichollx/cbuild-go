package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/rpnx/cbuild-go/app/cbuildapp"
)

func main() {
	if err := cbuildapp.CBuild.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
