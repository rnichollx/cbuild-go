package main

import (
	"fmt"
	"os"

	"gitlab.com/rpnx/cbuild-go/app/cbuildapp"
	"gitlab.com/rpnx/cbuild-go/app/csetupapp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <output-directory>\n", os.Args[0])
		os.Exit(1)
	}

	outDir := os.Args[1]
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	err = cbuildapp.CBuild.GenerateManpages(outDir)
	if err != nil {
		fmt.Printf("Error generating cbuild manpages: %v\n", err)
		os.Exit(1)
	}

	err = csetupapp.CSetup.GenerateManpages(outDir)
	if err != nil {
		fmt.Printf("Error generating csetup manpages: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Manpages generated in %s\n", outDir)
}
