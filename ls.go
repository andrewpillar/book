package main

import (
	"flag"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var LsCmd = &Command{
	Usage: "ls [-n] [-wc] [file]",
	Short: "list manuscripts and their chapters",
	Long: `List manuscripts in the current directory, if an argument is given, then this
will list the manuscript's chapters.

The -n flag will display chapter numbers, if listing a manuscript's chapters.

The -wc flag will print the word count of each manuscript, or each chapter if
an individual manuscript was given.`,
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
		names, err := filepath.Glob("*.mom")

		if err != nil {
			return err
		}

		books := make([]*Manuscript, 0, len(names))
		pad := 0

		for _, name := range names {
			ms, err := ParseManuscript(name)

			if err != nil {
				return err
			}

			title := ms.DocTitle()

			if l := utf8.RuneCountInString(title); l > pad {
				pad = l
			}
			books = append(books, ms)
		}

		for _, b := range books {
			title := b.DocTitle()

			if wc {
				if n := pad - utf8.RuneCountInString(title); n > 0 {
					title += strings.Repeat(" ", n)
				}

				cmd.Printf("%s %6s\n", title, formatNumber(b.WordCount()))
				continue
			}
			cmd.Println(title)
		}
		return nil
	}

	ms, err := ParseManuscript(args[0])

	if err != nil {
		return err
	}

	chapters, err := ms.Chapters()

	if err != nil {
		return err
	}

	pad := 0

	for _, ch := range chapters {
		title := ch.Title()

		if title == "" {
			title = ch.Number()
		}

		if l := utf8.RuneCountInString(title); l > pad {
			pad = l
		}
	}

	for i, ch := range chapters {
		if number {
			cmd.Printf("%3d ", i+1)
		}

		title := ch.Title()

		if title == "" {
			title = ch.Number()
		}
		cmd.Print(title)

		if wc {
			if n := pad - utf8.RuneCountInString(title); n > 0 {
				cmd.Print(strings.Repeat(" ", n))
			}
			cmd.Printf(" %6s", formatNumber(ch.WordCount()))
		}
		cmd.Println()
	}

	return nil
}
