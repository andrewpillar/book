package main

import (
	"bytes"
	"strings"
)

func PrintText(cmd *Command, s string) {
	sc := Scanner{
		Tokens: Tokenize(s),
	}

	tok := sc.Next()

	for tok != nil {
		if txt, ok := tok.(*Text); ok {
			cmd.Print(txt.Value)
		}
		tok = sc.Next()
	}
}

func PrintEpigraph(cmd *Command, sc *Scanner) {
loop:
	for {
		tok := sc.Next()

		if tok == nil {
			break
		}

		switch v := tok.(type) {
		case *Macro:
			if v.Name == "EPIGRAPH" {
				if v.Arg(0) == "OFF" {
					break loop
				}
			}
		case *Text:
			PrintText(cmd, v.Value)
			cmd.Println()
		}
	}
	cmd.Println()
}

func PrintParagraph(cmd *Command, sc *Scanner) {
	var buf bytes.Buffer

loop:
	for {
		tok := sc.Next()

		if tok == nil {
			break loop
		}

		switch v := tok.(type) {
		case *Macro:
			switch v.Name {
			case "DROPCAP":
				buf.WriteString(v.Arg(0))
			case "PP":
				sc.Back()
				break loop
			case "COLLATE":
				sc.Back()
				break loop
			}
		case *Text:
			buf.WriteString(v.Value)
			buf.WriteString(" ")
		}
	}

	if buf.String() == "" {
		return
	}
	PrintText(cmd, strings.TrimSuffix(buf.String(), " "))
}
