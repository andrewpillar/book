package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var PubCmd = &Command{
	Usage: "pub <-f docx|pdf> <file> [chapter,...]",
	Short: "publish the manuscript into a pdf",
	Long: `Publish the manuscript into the given format, either docx or pdf as specified via
the -f flag. If pdf, then groff is used under the hood to produce the final pdf.`,
	Run: pubCmd,
}

func pubCmd(cmd *Command, args []string) error {
	var format string

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&format, "f", "", "the format to publish in, either docx or pdf")
	fs.Parse(args)

	args = fs.Args()

	if len(args) == 0 {
		return ErrUsage
	}

	if format == "" {
		return ErrUsage
	}

	file := args[0]
	args = args[1:]

	ms, err := ParseManuscript(file)

	if err != nil {
		return err
	}

	// If chapters have been given, then make sure the manuscript only
	// contains that chapters we want to publish.
	if len(args) > 0 {
		chapters := ms.Chapters(args...)
		toks := make([]Token, 0, len(ms.Tokens))

		for _, tok := range ms.Tokens {
			if m, ok := tok.(*Macro); ok {
				if m.Name == "CHAPTER" || m.Name == "CHAPTER_TITLE" {
					break
				}
			}
			toks = append(toks, tok)
		}

		for _, ch := range chapters {
			toks = append(toks, ch.Tokens()...)
		}
		ms.Tokens = toks
	}

	// In the case of publishing a single chapter we want to remove the
	// trailing COLLATE macro. With this in place it will add an additional
	// blank page to the document, which we don't want.
	last := ms.Tokens[len(ms.Tokens)-1]

	if m, ok := last.(*Macro); ok {
		if m.Name == "COLLATE" {
			ms.Tokens = ms.Tokens[:len(ms.Tokens)-1]
		}
	}

	name := filepath.Base(file)
	name = name[:len(name)-4] + "." + format

	switch format {
	case "docx":
		docx, err := newDocxBuilder(name, ms)

		if err != nil {
			return err
		}

		if err := docx.build(); err != nil {
			return err
		}
	case "pdf":
		tmp, err := os.CreateTemp("", name)

		if err != nil {
			return err
		}

		defer os.Remove(tmp.Name())
		defer tmp.Close()

		if err := ms.WriteTo(tmp); err != nil {
			return err
		}

		f, err := os.Create(name)

		if err != nil {
			return err
		}

		defer f.Close()

		c := exec.Command("groff", "-k", "-mom", "-T", "pdf", tmp.Name())
		c.Stdin = os.Stdin
		c.Stdout = f
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			return err
		}
	default:
		return errors.New("unrecognized publish format")
	}

	fmt.Println(name)
	return nil
}
