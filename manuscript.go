package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Buffer struct {
	Buf  []byte
	Lit  []rune
	Pos  int
	EOF  int
	Line int
	Col  int
}

func BufferFile(name string) (*Buffer, error) {
	b, err := os.ReadFile(name)

	if err != nil {
		return nil, err
	}

	return &Buffer{
		Buf:  b,
		Pos:  0,
		EOF:  len(b),
		Line: 1,
	}, nil
}

func BufferString(s string) *Buffer {
	return &Buffer{
		Buf: []byte(s),
		EOF: len(s),
	}
}

func (b *Buffer) Get() rune {
	if b.Pos >= b.EOF {
		return -1
	}

	r := rune(b.Buf[b.Pos])
	w := 1

	if r >= utf8.RuneSelf {
		r, w = utf8.DecodeRune(b.Buf[b.Pos:])
	}

	b.Pos += w
	b.Col += w

	if r == '\n' {
		b.Line++
		b.Col = 0
	}
	return r
}

func (b *Buffer) GetLine() (string, bool) {
	r := b.Get()

	if r == -1 {
		return "", false
	}

	b.Lit = b.Lit[0:0]

	for r != '\n' {
		b.Lit = append(b.Lit, r)
		r = b.Get()
	}
	return string(b.Lit), true
}

func (b *Buffer) Seek(i int) {
	b.Pos = i

	if b.Pos < 0 {
		b.Pos = 0
	}
	if b.Pos >= b.EOF {
		b.Pos = b.EOF
	}
}

// Token represents the tokens parsed from the manuscript file.
type Token interface {
	aToken()

	// WriteTo writes the token as plain text to the given [io.Writer].
	// This should write it as the original plain text it was initially
	// parsed from.
	WriteTo(w io.Writer) error
}

// Macro represents a macro that has been parsed from the manuscript file.
type Macro struct {
	// Name is the macro name itself, either PP, or CHAPTER, etc.
	Name string

	// Args are the arguments given to the macro, if any.
	Args []string
}

func (m *Macro) aToken() {}

func (m *Macro) WriteTo(w io.Writer) error {
	if _, err := io.WriteString(w, "."+m.Name); err != nil {
		return err
	}

	if _, err := io.WriteString(w, " "); err != nil {
		return err
	}

	for _, arg := range m.Args {
		if _, err := fmt.Fprintf(w, "%q ", arg); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

// Text represents any line of text from the manuscript that is not a macro.
type Text struct {
	Value string
}

func (t *Text) aToken() {}

func (t *Text) WriteTo(w io.Writer) error {
	if _, err := io.WriteString(w, t.Value); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

// Inline represents an inline escape macro. This will not appear in the
// [Manuscript.Tokens] slice and is only used when generating the DOCX file of
// the manuscript when publishing.
type Inline struct {
	Escape string
}

func (i *Inline) aToken() {}

func (i *Inline) WriteTo(w io.Writer) error { return nil }

// Manuscript represents the parsed groff mom file. The file contents is parsed
// into a [Token] slice, which will either be of type [Macro] or [Text].
type Manuscript struct {
	Tokens []Token
}

// ParseManuscript parses a groff mom manuscript from the given file. The file
// is parsed line by line which is used to determine what is being parsed,
// whether it be a macro, or some plain text.
func ParseManuscript(name string) (*Manuscript, error) {
	buf, err := BufferFile(name)

	if err != nil {
		return nil, err
	}

	toks := make([]Token, 0)
	tmp := make([]rune, 0)

	for {
		line, ok := buf.GetLine()

		if !ok {
			break
		}

		if line != "" {
			if line[0] == '.' {
				var m Macro

				buf := BufferString(line)

				quoted := false
				r := buf.Get()

				for r != -1 {
					if r == '.' && m.Name == "" {
						r = buf.Get()

						for r != ' ' && r != '\t' && r != -1 {
							tmp = append(tmp, r)
							r = buf.Get()
						}

						m.Name = string(tmp)
						tmp = tmp[0:0]

						r = buf.Get()

						for r == ' ' || r == '\t' {
							r = buf.Get()
						}
						continue
					}

					if r == '"' {
						quoted = !quoted
						r = buf.Get()
						continue
					}

					if r == ' ' && !quoted {
						m.Args = append(m.Args, string(tmp))
						tmp = tmp[0:0]

						r = buf.Get()
						continue
					}

					tmp = append(tmp, r)
					r = buf.Get()
				}

				if len(tmp) > 0 {
					m.Args = append(m.Args, string(tmp))
					tmp = tmp[0:0]
				}

				toks = append(toks, &m)
				continue
			}
		}

		toks = append(toks, &Text{
			Value: line,
		})
	}

	return &Manuscript{
		Tokens: toks,
	}, nil
}

// Macro returns the first macro by the given name. This should be used for
// macros that will only appear once in a manuscript, such as DOCTITLE or
// AUTHOR.
func (ms *Manuscript) Macro(name string) *Macro {
	for _, tok := range ms.Tokens {
		if m, ok := tok.(*Macro); ok {
			if m.Name == name {
				return m
			}
		}
	}
	return nil
}

// Get returns the first argument of the first macro by the given name. This
// should be used for macros that will only appear once in a manuscript, such as
// DOCTITLE or AUTHOR.
func (ms *Manuscript) Get(macro string) string {
	if m := ms.Macro(macro); m != nil {
		if len(m.Args) > 0 {
			return m.Args[0]
		}
	}
	return ""
}

// DocTitle returns the title from DOCTITLE.
func (ms *Manuscript) DocTitle() string {
	return ms.Get("DOCTITLE")
}

// Author returns the author from AUTHOR.
func (ms *Manuscript) Author() string {
	return ms.Get("AUTHOR")
}

// Copyright returns the copyright from COPYRIGHT.
func (ms *Manuscript) Copyright() string {
	return ms.Get("COPYRIGHT")
}

// PrintStyle returns the print style from PRINTSTYLE.
func (ms *Manuscript) PrintStyle() string {
	return ms.Get("PRINTSTYLE")
}

// Chapter represents a single chapter within a manuscript.
//
// The Count of the chapter is the numeric count of the chapter itself. That
// is, whether it is the first, or second, etc. chapter in the manuscript. We
// call this count as opposed to number because the chapter number is something
// that can be specified via CHAPTER, which can be a number, a numeral, or a
// word.
//
// Start and End mark the positions within the [Manuscript.Tokens] slice as to
// where the chapter's contents is.
type Chapter struct {
	*Manuscript

	Count int
	Start int
	End   int
}

// Tokens returns the slice of tokens that make up the contents of the chapter
// within the manuscript.
func (ch *Chapter) Tokens() []Token {
	return ch.Manuscript.Tokens[ch.Start:ch.End]
}

// Number returns the number of the chapter as specified via CHAPTER.
func (ch *Chapter) Number() string {
	number := ""

	for _, tok := range ch.Tokens() {
		if m, ok := tok.(*Macro); ok {
			if m.Name == "CHAPTER" {
				if len(m.Args) > 0 {
					number = "Chapter " + m.Args[0]
				}
			}
		}
	}
	return number
}

// Title returns the title of the chapter as specified via CHAPTER_TITLE.
func (ch *Chapter) Title() string {
	title := ""

	for _, tok := range ch.Tokens() {
		if m, ok := tok.(*Macro); ok {
			if m.Name == "CHAPTER_TITLE" {
				if len(m.Args) > 0 {
					title = m.Args[0]
				}
			}
		}
	}
	return title
}

// WordCount returns the word count of all the text content within the chapter.
func (ch *Chapter) WordCount() int {
	wc := 0

	for _, tok := range ch.Tokens() {
		if txt, ok := tok.(*Text); ok {
			if txt.Value == "" || txt.Value == " " {
				continue
			}
			wc += len(strings.Split(strings.TrimSpace(txt.Value), " "))
		}
	}
	return wc
}

// Chapters returns a slice of the chapters within the manuscript. The chapters
// returned can be limited via names, which can either include the chapter
// titles or the chapter counts themselves. For example, passing "1", and "2"
// would return a slice of the first two chapters.
func (ms *Manuscript) Chapters(names ...string) []*Chapter {
	set := make(map[string]struct{})

	for _, name := range names {
		set[name] = struct{}{}
	}

	chapters := make([]*Chapter, 0)
	count := 0

	sc := scanner{
		toks: ms.Tokens,
	}

	for {
		tok := sc.next()

		if tok == nil {
			break
		}

		if m, ok := tok.(*Macro); ok {
			ch := Chapter{
				Manuscript: ms,
			}

			if m.Name == "CHAPTER" {
				count++

				ch.Count = count
				ch.Start = sc.pos

				tok = sc.next()

				// Parse the rest of the chapter, which would include the
				// CHAPTER_TITLE macro if also specified. We don't want to
				// mistakenly count a chapter twice.
				goto parseChapter
			}

			if m.Name == "CHAPTER_TITLE" {
				count++

				ch.Count = count
				ch.Start = sc.pos

				tok = sc.next()
			}

		parseChapter:
			for {
				if tok == nil {
					ch.End = sc.pos
					break
				}

				if m, ok := tok.(*Macro); ok {
					if m.Name == "COLLATE" {
						ch.End = sc.pos
						break
					}
				}
				tok = sc.next()
			}

			// Filter out the chapters that have been specified, if any. First we check
			// for chapter counts, then fallback to checking chapter titles.
			if len(names) > 0 {
				scount := strconv.Itoa(count)

				if _, ok := set[scount]; !ok {
					title := ch.Title()

					if _, ok := set[title]; ok {
						continue
					}
				}
			}
			chapters = append(chapters, &ch)
		}
	}
	return chapters
}

// WriteTo writes the contents of the entire manuscript to the given writer.
// This will produce a 1-to-1 of what is on disk from the original groff mom
// manuscript file.
func (ms *Manuscript) WriteTo(w io.Writer) error {
	for _, tok := range ms.Tokens {
		if err := tok.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}
