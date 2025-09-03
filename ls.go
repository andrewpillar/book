package main

import (
	"flag"
	"fmt"
	"strings"
	"unicode/utf8"
)

var LsCmd = &Command{
	Usage: "ls [-n] [-wc] <file>",
	Short: "list chapters",
	Long: `List the chapters in the given manuscript file.

The -n flag will display chapter numbers.

The -wc flag will print the word count of each chapter.`,
	Run: lsCmd,
}

func lsCmd(cmd *Command, args []string) error {
	var (
		number bool
		wc     bool
	)

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.BoolVar(&number, "n", false, "display chapter number")
	fs.BoolVar(&wc, "wc", false, "display word count of the chapters")
	fs.Parse(args)

	args = fs.Args()

	if len(args) == 0 {
		return ErrUsage
	}

	ms, err := ParseManuscript(args[0])

	if err != nil {
		return err
	}

	chapters := ms.Chapters()
	pad := 0

	for _, ch := range chapters {
		if l := utf8.RuneCountInString(ch.Title()); l > pad {
			pad = l
		}
	}

	for _, ch := range chapters {
		if number {
			fmt.Printf("%3d ", ch.Number)
		}

		title := ch.Title()
		fmt.Print(title)

		if wc {
			if n := pad - utf8.RuneCountInString(title); n > 0 {
				fmt.Print(strings.Repeat(" ", n))
			}
			fmt.Printf(" %6s", formatNumber(ch.WordCount()))
		}
		fmt.Println()
	}

	return nil
}
