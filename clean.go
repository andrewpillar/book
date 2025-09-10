package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var CleanCmd = &Command{
	Usage: "clean",
	Short: "remove pdf files for manuscripts",
	Long: `Remove the DOCX and PDF files for the published manuscripts.

The -v flag will print out the names of each file deleted.`,
	Run: cleanCmd,
}

func cleanCmd(cmd *Command, args []string) error {
	var verbose bool

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.BoolVar(&verbose, "v", false, "print out the cleaned up files")
	fs.Parse(args)

	names, err := filepath.Glob("*.mom")

	if err != nil {
		return err
	}

	for _, name := range names {
		name = name[:len(name)-4]

		pdfMatches, err := filepath.Glob(name + "*.pdf")

		if err != nil {
			return err
		}

		docxMatches, err := filepath.Glob(name + "*.docx")

		if err != nil {
			return err
		}

		matches := append(docxMatches, pdfMatches...)

		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}

			if verbose {
				fmt.Println(match)
			}
		}
	}
	return nil
}
