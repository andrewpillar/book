package main

import (
	"bytes"
	"strings"

	"github.com/gomutex/godocx"
	"github.com/gomutex/godocx/docx"
	"github.com/gomutex/godocx/wml/ctypes"
	"github.com/gomutex/godocx/wml/stypes"
)

type docxBuilder struct {
	name     string
	font     string
	color    string
	ms       *Manuscript
	chapters []string
	doc      *docx.RootDoc
}

// newDocxBuilder returns a docxBuilder for building the given [Manuscript]
// into a DOCX file of the given name. The font and color for the document will
// be Times New Roman and #000000 respectively.
func newDocxBuilder(name string, ms *Manuscript, chapters ...string) (*docxBuilder, error) {
	doc, err := godocx.NewDocument()

	if err != nil {
		return nil, err
	}

	return &docxBuilder{
		name:     name,
		font:     "Times New Roman",
		color:    "#000000",
		ms:       ms,
		chapters: chapters,
		doc:      doc,
	}, nil
}

// paragraphProp returns the [ctypes.ParagaphProp] for a paragraph, this will
// be configured with the appropriate line height, which will be double spacing
// or close to.
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

// defaultRun returns the [docx.Run] to be applied to a paragraph, this will
// ensure the font and color will match the default, and set the font size to
// the given size.
func (b *docxBuilder) defaultRun(r *docx.Run, size uint64) *docx.Run {
	r.Font(b.font)
	r.Color(b.color)
	r.Size(size)

	return r
}

func (b *docxBuilder) buildItalics(p *docx.Paragraph, sc *Scanner, size uint64) {
	var buf bytes.Buffer

loop:
	for {
		tok := sc.Next()

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
	sc := Scanner{
		Tokens: Tokenize(txt),
	}

	for {
		tok := sc.Next()

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

func (b *docxBuilder) buildEpigraph(sc *Scanner) {
loop:
	for {
		tok := sc.Next()

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

func (b *docxBuilder) buildParagraph(sc *Scanner, indent bool) {
	p := b.doc.AddParagraph("")
	p.GetCT().Property = b.paragraphProp()

	start := true

	var buf bytes.Buffer

loop:
	for {
		tok := sc.Next()

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
				sc.Back()
				break loop
			}
		case *Text:
			// There could be multiple Text tokens to a paragraph,
			// therefore only indent if this is the first Text token
			// and if indent is set to true.
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
	if number := ch.Number(); number != "" {
		hdr, err := b.doc.AddHeading("", 1)

		if err != nil {
			return err
		}

		line := 0

		hdr.GetCT().Property = b.paragraphProp()
		hdr.GetCT().Property.Spacing = &ctypes.Spacing{
			Line: &line,
		}
		hdr.Justification(stypes.JustificationCenter)

		run := hdr.AddText(number)

		run = b.defaultRun(run, 18)
		run.Bold(true)
	}

	if title := ch.Title(); title != "" {
		hdr, err := b.doc.AddHeading("", 1)

		if err != nil {
			return err
		}

		hdr.GetCT().Property = b.paragraphProp()
		hdr.Justification(stypes.JustificationCenter)

		run := hdr.AddText(title)

		run = b.defaultRun(run, 18)
		run.Bold(true)
		run.Italic(true)
	}

	sc := Scanner{
		Tokens: ch.Tokens,
	}

	indent := false

	for {
		tok := sc.Next()

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

	// Precede with blank paragraphs so that the title is centered.
	for i := 0; i < 10; i++ {
		b.doc.AddParagraph("")
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

	chapters, err := b.ms.Chapters()

	if err != nil {
		return err
	}

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
