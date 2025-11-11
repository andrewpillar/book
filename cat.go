package main

import (
	"bytes"
	"fmt"
	"strings"
)

var CatCmd = &Command{
	Usage: "cat <file> [chapter,...]",
	Short: "print out text content of the manuscript",
	Run:   catCmd,
}

func printText(txt string) {
	sc := scanner{
		toks: tokenize(txt),
	}

	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		if txt, ok := tok.(*Text); ok {
			fmt.Print(txt.Value)
		}
	}
}

func printEpigraph(sc *scanner) {
loop:
	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		switch v := tok.(type) {
		case *Macro:
			if v.Name == "EPIGRAPH" {
				if len(v.Args) > 0 && v.Args[0] == "OFF" {
					break loop
				}
			}
		case *Text:
			printText(v.Value)
		}
	}

	fmt.Println()
	fmt.Println()
}

func printParagraph(sc *scanner, indent bool) {
	start := true

	var buf bytes.Buffer

loop:
	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		switch v := tok.(type) {
		case *Macro:
			switch v.Name {
			case "DROPCAP":
				if len(v.Args) > 0 {
					buf.WriteString(v.Args[0])
				}
			case "PP":
				sc.back()
				break loop
			}
		case *Text:
			if start && indent {
				buf.WriteString("        ")
				start = false
			}
			buf.WriteString(v.Value)
			buf.WriteString(" ")
		}
	}

	buf.WriteString("\n")

	printText(strings.TrimSuffix(buf.String(), " "))
}

func printChapter(ch *Chapter) {
	if number := ch.Number(); number != "" {
		fmt.Println(number)
	}
	fmt.Println(ch.Title())
	fmt.Println()

	sc := scanner{
		toks: ch.Tokens,
	}

	indent := false

	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		if m, ok := tok.(*Macro); ok {
			switch m.Name {
			case "EPIGRAPH":
				printEpigraph(&sc)
			case "PP":
				printParagraph(&sc, indent)
				indent = true
			}
		}
	}
}

func catCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	file := args[0]
	args = args[1:]

	ms, err := ParseManuscript(file)

	if err != nil {
		return err
	}

	chapters := ms.Chapters(args...)

	for i, ch := range chapters {
		printChapter(ch)

		if i != len(chapters)-1 {
			fmt.Println()
		}
	}
	return nil
}
