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

type Token interface {
	aToken()

	WriteTo(w io.Writer) error
}

type Macro struct {
	Name string
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

type Inline struct {
	Escape string
}

func (i *Inline) aToken() {}

func (i *Inline) WriteTo(w io.Writer) error { return nil }

type Manuscript struct {
	Tokens []Token
}

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

func (ms *Manuscript) Get(macro string) string {
	if m := ms.Macro(macro); m != nil {
		if len(m.Args) > 0 {
			return m.Args[0]
		}
	}
	return ""
}

func (ms *Manuscript) DocTitle() string {
	return ms.Get("DOCTITLE")
}

func (ms *Manuscript) Author() string {
	return ms.Get("AUTHOR")
}

func (ms *Manuscript) Copyright() string {
	return ms.Get("COPYRIGHT")
}

func (ms *Manuscript) PrintStyle() string {
	return ms.Get("PRINTSTYLE")
}

type Chapter struct {
	*Manuscript

	Number int
	Start  int
	End    int
}

func (ch *Chapter) Tokens() []Token {
	return ch.Manuscript.Tokens[ch.Start:ch.End]
}

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

	if title == "" {
		title = strconv.Itoa(ch.Number)
	}
	return title
}

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

func (ms *Manuscript) Chapters(names ...string) []*Chapter {
	set := make(map[string]struct{})

	for _, name := range names {
		set[name] = struct{}{}
	}

	chapters := make([]*Chapter, 0)
	num := 0

	for i, tok := range ms.Tokens {
		if m, ok := tok.(*Macro); ok {
			if m.Name == "CHAPTER" || m.Name == "CHAPTER_TITLE" {
				num++

				// Determine if we actually want this chapter returned in the slice.
				if len(names) > 0 {
					title := strconv.Itoa(num)

					// First check to see the chapter number itself has been
					// given. If not, fallback to the chapter title itself.
					if _, ok := set[title]; !ok {
						if m.Name == "CHAPTER_TITLE" && len(m.Args) > 0 {
							title = m.Args[0]
						}

						if _, ok := set[title]; !ok {
							continue
						}
					}
				}

				ch := Chapter{
					Manuscript: ms,
					Number:     num,
					Start:      i,
				}

				for j, tok := range ms.Tokens[ch.Start:] {
					if m, ok := tok.(*Macro); ok {
						if m.Name == "COLLATE" {
							ch.End = i + j + 1
							break
						}
					}

					// If the end has not been set because there is no COLLATE
					// then assume this is the last chapter and EOF, so set the
					// end to the end of the token slice.
					if j == len(ms.Tokens[ch.Start:])-1 {
						ch.End = i + j + 1
					}
				}
				chapters = append(chapters, &ch)
			}
		}
	}
	return chapters
}

func (ms *Manuscript) WriteTo(w io.Writer) error {
	for _, tok := range ms.Tokens {
		if err := tok.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}
