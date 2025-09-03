package main

import (
	"bytes"
	"strings"

	"github.com/gomutex/godocx"
	"github.com/gomutex/godocx/docx"
	"github.com/gomutex/godocx/wml/ctypes"
	"github.com/gomutex/godocx/wml/stypes"
)

type scanner struct {
	pos  int
	toks []Token
}

func (sc *scanner) back() {
	sc.pos--

	if sc.pos < 0 {
		sc.pos = 0
	}
}

func (sc *scanner) next() Token {
	if sc.pos >= len(sc.toks) {
		return nil
	}

	tok := sc.toks[sc.pos]
	sc.pos++

	return tok
}

type docxBuilder struct {
	name  string
	font  string
	color string
	ms    *Manuscript
	doc   *docx.RootDoc
}

func newDocxBuilder(name string, ms *Manuscript) (*docxBuilder, error) {
	doc, err := godocx.NewDocument()

	if err != nil {
		return nil, err
	}

	return &docxBuilder{
		name:  name,
		font:  "Times New Roman",
		color: "#000000",
		ms:    ms,
		doc:   doc,
	}, nil
}

func (b *docxBuilder) paragraphProp() *ctypes.ParagraphProp {
	line := 500
	after := uint64(0)

	prop := ctypes.DefaultParaProperty()
	prop.Spacing = &ctypes.Spacing{
		Line:  &line,
		After: &after,
	}
	return prop
}

func (b *docxBuilder) tokenize(txt string) []Token {
	buf := BufferString(txt)
	tmp := make([]rune, 0, len(txt))

	toks := make([]Token, 0)

	r := buf.Get()

	for r != -1 {
		if r == '\\' {
			toks = append(toks, &Text{
				Value: string(tmp),
			})

			tmp = tmp[0:0]

			r = buf.Get()

			if r == '*' || r == '[' {
				r = buf.Get()

				if r == '[' {
					r = buf.Get()
				}

				for r != ']' {
					tmp = append(tmp, r)
					r = buf.Get()
				}

				toks = append(toks, &Inline{
					Escape: string(tmp),
				})

				tmp = tmp[0:0]
				r = buf.Get()
				continue
			}
		}

		tmp = append(tmp, r)
		r = buf.Get()
	}

	if len(tmp) > 0 {
		toks = append(toks, &Text{
			Value: string(tmp),
		})
	}
	return toks
}

func (b *docxBuilder) defaultRun(r *docx.Run, size uint64) *docx.Run {
	r.Font(b.font)
	r.Color(b.color)
	r.Size(size)

	return r
}

func (b *docxBuilder) buildItalics(p *docx.Paragraph, sc *scanner, size uint64) {
	var buf bytes.Buffer

loop:
	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		switch v := tok.(type) {
		case *Inline:
			if v.Escape == "PREV" {
				break loop
			}
		case *Text:
			buf.WriteString(v.Value)
		}
	}
	b.defaultRun(p.AddText(buf.String()), size).Italic(true)
}

func (b *docxBuilder) buildText(p *docx.Paragraph, txt string, size uint64) {
	sc := scanner{
		toks: b.tokenize(txt),
	}

	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		switch v := tok.(type) {
		case *Inline:
			switch v.Escape {
			case "IT":
				b.buildItalics(p, &sc, size)
			case "lq":
				fallthrough
			case "rq":
				b.defaultRun(p.AddText("\""), size)
			}
		case *Text:
			b.defaultRun(p.AddText(v.Value), size)
		}
	}
}

func (b *docxBuilder) buildEpigraph(sc *scanner) {
loop:
	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		switch v := tok.(type) {
		case *Macro:
			switch v.Name {
			case "EPIGRAPH":
				if len(v.Args) > 0 && v.Args[0] == "OFF" {
					break loop
				}
			}
		case *Text:
			p := b.doc.AddParagraph("")
			p.GetCT().Property = b.paragraphProp()
			p.Justification(stypes.JustificationCenter)
			b.buildText(p, v.Value, 10)
		}
	}
}

func (b *docxBuilder) buildParagraph(sc *scanner, indent bool) {
	p := b.doc.AddParagraph("")
	p.GetCT().Property = b.paragraphProp()

	// Denotes the start of the paragraph, used to ensure the paragraph
	// is indented.
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
			case "RIGHT":
				p.Justification(stypes.JustificationRight)
			case "PP":
				sc.back()
				break loop
			}
		case *Text:
			if start && indent {
				buf.WriteString("\t")
				start = false
			}
			buf.WriteString(v.Value)
			buf.WriteString(" ")
		}
	}
	b.buildText(p, strings.TrimSuffix(buf.String(), " "), 12)
}

func (b *docxBuilder) buildChapter(ch *Chapter) error {
	title, err := b.doc.AddHeading("", 1)

	if err != nil {
		return err
	}

	title.GetCT().Property = b.paragraphProp()
	title.Justification(stypes.JustificationCenter)

	run := title.AddText(ch.Title())

	run = b.defaultRun(run, 18)
	run.Bold(true)
	run.Italic(true)

	sc := scanner{
		toks: ch.Tokens(),
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
				b.buildEpigraph(&sc)
			case "PP":
				b.buildParagraph(&sc, indent)
				indent = true
			}
		}
	}
	return nil
}

func (b *docxBuilder) buildTitlePage() error {
	b.doc.Document.Body.SectPr.TitlePg = &ctypes.GenSingleStrVal[stypes.OnOff]{
		Val: stypes.OnOffOn,
	}

	title, err := b.doc.AddHeading("", 0)

	if err != nil {
		return err
	}

	title.Justification(stypes.JustificationCenter)

	after := uint64(0)

	ct := title.GetCT()
	ct.Property.Border = &ctypes.ParaBorder{
		Bottom: &ctypes.Border{},
	}
	ct.Property.Spacing = &ctypes.Spacing{
		After: &after,
	}

	run := title.AddText(b.ms.DocTitle())

	run = b.defaultRun(run, 18)
	run.Bold(true)
	run.Italic(true)

	for _, s := range []string{"by", b.ms.Author()} {
		p := b.doc.AddParagraph("")
		p.Justification(stypes.JustificationCenter)
		p.GetCT().Property.Spacing = &ctypes.Spacing{
			After: &after,
		}

		run := p.AddText(s)

		run = b.defaultRun(run, 12)
		run.Italic(true)
	}

	b.doc.AddPageBreak()

	return nil
}

func (b *docxBuilder) build() error {
	margin := 1300

	b.doc.Document.Body.SectPr.PageMargin.Left = &margin
	b.doc.Document.Body.SectPr.PageMargin.Right = &margin

	if err := b.buildTitlePage(); err != nil {
		return err
	}

	chapters := b.ms.Chapters()

	for i, ch := range chapters {
		if err := b.buildChapter(ch); err != nil {
			return err
		}

		if i != len(chapters)-1 {
			b.doc.AddPageBreak()
		}
	}

	if err := b.doc.SaveTo(b.name); err != nil {
		return err
	}
	return nil
}
