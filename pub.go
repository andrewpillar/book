package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var PubCmd = &Command{
	Usage: "pub <-f docx|pdf> <-wc count> <file> [chapter,...]",
	Short: "publish the manuscript into a pdf",
	Long: `Publish the manuscript into the given format, either docx or pdf as specified via
the -f flag. If pdf, then groff is used under the hood to produce the final pdf.`,
	Run: pubCmd,
}

func pubCmd(cmd *Command, args []string) error {
	var (
		format string
		wc     int
	)

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&format, "f", "", "the format to publish in, either docx or pdf")
	fs.IntVar(&wc, "wc", 0, "the number of words to publish")
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

	name := filepath.Base(file)
	name = name[:len(name)-4]

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

		if len(chapters) > 1 {
			name += "-chapters-"
			name += strconv.Itoa(chapters[0].Count) + "-to-"
			name += strconv.Itoa(chapters[len(chapters)-1].Count)
		} else {
			name += "-chapter-" + strconv.Itoa(chapters[0].Count)
		}

		ms.Tokens = toks
	}

	if wc > 0 {
		sum := 0
		pos := 0

		for i, tok := range ms.Tokens {
			if sum >= wc {
				pos = i
				break
			}

			if txt, ok := tok.(*Text); ok {
				if txt.Value == "" || txt.Value == " " {
					continue
				}
				sum += len(strings.Split(strings.TrimSpace(txt.Value), " "))
			}
		}

		name += "-first-" + strconv.Itoa(wc) + "-words"

		ms.Tokens = ms.Tokens[:pos]
	}

	name += "." + format

	// In the case of publishing a single chapter we want to remove the
	// trailing COLLATE macro. With this in place it will add an additional
	// blank page to the document, which we don't want.
	last := ms.Tokens[len(ms.Tokens)-1]

	if m, ok := last.(*Macro); ok {
		if m.Name == "COLLATE" {
			ms.Tokens = ms.Tokens[:len(ms.Tokens)-1]
		}
	}

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
