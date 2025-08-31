package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/gomutex/godocx"
	"github.com/gomutex/godocx/wml/stypes"
)

var PubCmd = &Command{
	Usage: "pub <-f docx|pdf> <file> [chapter,...]",
	Short: "publish the manuscript into a pdf",
	Long: `Publish the manuscript into a PDF. This will use groff under the hood to print
out the PDF.

The -f flag specifies the format into which the manuscript should be published.
This accepts either docx or pdf.`,
	Run: pubCmd,
}

//go:embed manuscript.tmpl
var manuscriptTmpl []byte

func formatNumber(n int) string {
	s := strconv.FormatInt(int64(n), 10)

	for i := len(s); i > 3; {
		i -= 3
		s = s[:i] + "," + s[i:]
	}
	return s
}

func ManuscriptTemplate() (*template.Template, error) {
	tmpl := template.New("manuscript.tmpl")
	tmpl.Funcs(template.FuncMap{
		"format_number": formatNumber,
		"now":           time.Now,
	})

	return tmpl.Parse(string(manuscriptTmpl))
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

	var chapters []string

	if len(args) > 1 {
		chapters = args[1:]
	}

	ms, err := LoadManuscript(file)

	if err != nil {
		return err
	}

	if len(chapters) > 0 {
		only := make(map[string]struct{})

		for _, title := range chapters {
			only[title] = struct{}{}
		}

		keep := make([]*Chapter, 0, len(ms.Chapters))

		for _, ch := range ms.Chapters {
			if _, ok := only[ch.Title]; ok {
				keep = append(keep, ch)
			}
		}
		ms.Chapters = keep
	}

	name := filepath.Base(file)
	name = name[:len(name)-4]

	switch format {
	case "docx":
		doc, err := godocx.NewDocument()

		if err != nil {
			return err
		}

		title, err := doc.AddHeading(ms.Title, 0)

		if err != nil {
			return err
		}

		title.Justification(stypes.JustificationCenter)

		doc.AddParagraph("").AddText("by").Italic(true)
		doc.AddParagraph("").AddText(ms.Author).Italic(true)

		doc.AddPageBreak()

		for _, ch := range ms.Chapters {
			title, err := doc.AddHeading(ch.Title, 1)

			if err != nil {
				return err
			}

			title.Justification(stypes.JustificationCenter)

			for i, p := range ch.Paragraphs() {
				if i != 0 {
					p = "\t" + p
				}
				doc.AddParagraph(p)
			}

			doc.AddPageBreak()
		}

		if err := doc.SaveTo(name + ".docx"); err != nil {
			return err
		}
		fmt.Println(name + ".docx")
	case "pdf":
		tmp, err := os.CreateTemp("", name)

		if err != nil {
			return err
		}

		defer os.Remove(tmp.Name())
		defer tmp.Close()

		if _, err := io.Copy(tmp, ms.Buffer); err != nil {
			return err
		}

		f, err := os.Create(name + ".pdf")

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
		fmt.Println(f.Name())
	default:
		return errors.New("unrecognized publish format")
	}
	return nil
}
