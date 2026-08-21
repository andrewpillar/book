package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
)

var PubCmd = &Command{
	Usage: "pub <-f docx|pdf> <-wc count> <-o file> <-v> <file> [chapter,...]",
	Short: "publish the manuscript into a pdf or docx file",
	Long: `Publish the manuscript as the given format, either docx or pdf as specified via
the -f flag. If pdf, then pdfmom is used under the hood to produce the final
pdf.

The -wc flag can be given to only publish the first N words of the manuscript.
If given alongside a chapter, then the word count limit will be applied from
that chapter onwards.

The -o flag can be given to control the output name of the file. By default the
output name of the final file will be the name of the manuscript, suffixed with
the format, either pdf or docx.

The -o flag takes placeholder strings to better control the formatting of the
filename,

    %T - The title of the manuscript
    %A - The author of the manuscript
`,

	Run: pubCmd,
}

func pubCmd(cmd *Command, args []string) error {
	var (
		format  string
		wc      int
		out     string
		verbose bool
	)

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&format, "f", "", "the format to publish in, either docx or pdf")
	fs.IntVar(&wc, "wc", 0, "the number of words to publish")
	fs.StringVar(&out, "o", "", "write to file instead of the default")
	fs.BoolVar(&verbose, "v", false, "print the name of the file once published")
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
		chapters, err := ms.Chapters(args...)

		if err != nil {
			return err
		}

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
			toks = append(toks, ch.Tokens...)
		}
		ms.Tokens = toks
	}

	name := file[:len(file)-4] + "." + format

	if out != "" {
		var namebuf bytes.Buffer

		outbuf := BufferString(out)

		r := outbuf.Get()

	loop:
		for r != -1 {
			if r == '%' {
				switch outbuf.Get() {
				case -1:
					break loop
				case 'T':
					namebuf.WriteString(ms.DocTitle())
				case 'A':
					namebuf.WriteString(ms.Author())
				}

				r = outbuf.Get()
				continue
			}

			namebuf.WriteRune(r)
			r = outbuf.Get()
		}
		name = namebuf.String() + "." + format
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
				sum += len(txt.Words())
			}
		}
		ms.Tokens = ms.Tokens[:pos]
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

	dir := filepath.Dir(file)

	// Change into the directory of the source file being published. This is done
	// to ensure that relative paths used via PDF_IMAGE, and such other macros,
	// do not cause errors when being processed into PDF.
	if err := os.Chdir(dir); err != nil {
		return err
	}

	switch format {
	case "docx":
		if err := os.Remove(name); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		if err := WriteToDOCX(name, ms); err != nil {
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

		c := exec.Command("pdfmom", "-k", tmp.Name())
		c.Stdin = os.Stdin
		c.Stdout = f
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			return err
		}
	default:
		return errors.New("unrecognized publish format, must be one of: [docx, pdf]")
	}

	if verbose {
		cmd.Println(name)
	}
	return nil
}
