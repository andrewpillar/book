package main

import (
	"errors"
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

// Token represents the tokens parsed from the manuscript file.
type Token interface {
	aToken()

	// WriteTo writes the token as plain text to the given [io.Writer].
	// This should write it as the original plain text it was initially
	// parsed from.
	WriteTo(w io.Writer) error
}

// Scanner allows for iterating over a slice of Tokens, with support for back
// tracking too.
type Scanner struct {
	Pos    int
	Tokens []Token
}

func (sc *Scanner) Back() {
	sc.Pos--

	if sc.Pos < 0 {
		sc.Pos = 0
	}
}

func (sc *Scanner) Next() Token {
	if sc.Pos >= len(sc.Tokens) {
		return nil
	}

	tok := sc.Tokens[sc.Pos]
	sc.Pos++

	return tok
}

func (sc *Scanner) Peek() Token {
	tok := sc.Next()
	sc.Back()
	return tok
}

// Macro represents a macro that has been parsed from the manuscript file.
type Macro struct {
	// Raw is the original macro text that was parsed.
	Raw []rune

	// Name is the macro name itself, either PP, or CHAPTER, etc.
	Name string

	// Args are the arguments given to the macro, if any.
	Args []string
}

func (m *Macro) aToken() {}

func (m *Macro) Arg(n int) string {
	if len(m.Args) == 0 {
		return ""
	}

	if n >= len(m.Args) {
		return ""
	}
	return m.Args[n]
}

func (m *Macro) WriteTo(w io.Writer) error {
	if _, err := fmt.Fprintln(w, string(m.Raw)); err != nil {
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

// Tokenize splits up the given string into a slice of tokens. The tokens in the
// slice will either be of type [Text] or [Inline]. This is used to ensure that
// inline escape macros for can be used to format the text appropriately for the
// DOCX format, such as italics.
//
// This doesn't do validation and assumes that input text is "correct". So if an
// inline escape macro is closed off properly, the manuscript will have broken
// formatting. This is consistent with how groff works anyway.
func Tokenize(txt string) []Token {
	buf := BufferString(txt)

	// tmp is used to store the part of the string we have scanned in so
	// far.
	tmp := make([]rune, 0, len(txt))

	toks := make([]Token, 0)

	r := buf.Get()

	for r != -1 {
		if r == '\\' {
			// Found the start of an inline macro, so tokenize what we have in the tmp
			// buffer before we parse the inline macro.
			if s := string(tmp); s != "" {
				toks = append(toks, &Text{
					Value: string(tmp),
				})
			}

			tmp = tmp[0:0]

			r = buf.Get()

			// We're at the end of the inline escape macro, so parse the name of macro
			// we have.
			if r == '*' || r == '[' {
				r = buf.Get()

				if r == '[' {
					r = buf.Get()
				}

				for r != ']' {
					tmp = append(tmp, r)
					r = buf.Get()
				}

				// Tokenize.
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
		// Tokenize whatever text is remaining in the tmp buffer.
		toks = append(toks, &Text{
			Value: string(tmp),
		})
	}
	return toks
}

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
				m := Macro{
					Raw: []rune(line),
				}

				buf := BufferString(line)
				buf.Get()

				quoted := false
				r := buf.Get()

				// Skip over white space, it is valid to have a space between the leading . and
				// the name of the macro.
				for r == ' ' || r == '\t' {
					r = buf.Get()
				}

				for r != ' ' && r != '\t' && r != -1 {
					tmp = append(tmp, r)
					r = buf.Get()
				}

				m.Name = string(tmp)
				tmp = tmp[0:0]

				for r == ' ' || r == '\t' {
					r = buf.Get()
				}

				for r != -1 {
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
type Chapter struct {
	*Manuscript

	Count  int
	Tokens []Token
}

// Number returns the number of the chapter as specified via CHAPTER.
func (ch *Chapter) Number() string {
	number := ""

	for _, tok := range ch.Tokens {
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

	for _, tok := range ch.Tokens {
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

	for _, tok := range ch.Tokens {
		if txt, ok := tok.(*Text); ok {
			if txt.Value == "" || txt.Value == " " {
				continue
			}
			wc += len(strings.Split(strings.TrimSpace(txt.Value), " "))
		}
	}
	return wc
}

// WordCount returns the word count of all the text content within the
// manuscript.
func (ms *Manuscript) WordCount() int {
	wc := 0

	for _, tok := range ms.Tokens {
		if txt, ok := tok.(*Text); ok {
			if txt.Value == "" || txt.Value == " " {
				continue
			}
			wc += len(strings.Split(strings.TrimSpace(txt.Value), " "))
		}
	}
	return wc
}

var (
	ErrRangeFormat  = errors.New("range must be in format of start:end")
	ErrRangeType    = errors.New("range can only contain integers")
	ErrRangeInvalid = errors.New("range invalid start must be less than end")
)

// Chapters returns a slice of the chapters within the manuscript, as specified
// by the given names. The names can either be chapter titles, numbers, or a
// range of numbers. For example "1:4", would return chapters 1 through to 4.
func (ms *Manuscript) Chapters(names ...string) ([]*Chapter, error) {
	set := make(map[string]struct{})

	for _, name := range names {
		if strings.Contains(name, ":") {
			parts := strings.Split(name, ":")

			if len(parts) != 2 {
				return nil, ErrRangeFormat
			}

			start, err := strconv.Atoi(parts[0])

			if err != nil {
				return nil, ErrRangeType
			}

			end, err := strconv.Atoi(parts[1])

			if err != nil {
				return nil, ErrRangeType
			}

			if start > end {
				return nil, ErrRangeInvalid
			}

			for i := start; i < end+1; i++ {
				set[strconv.Itoa(i)] = struct{}{}
			}
			continue
		}
		set[name] = struct{}{}
	}

	chapters := make([]*Chapter, 0)
	count := 0

	sc := Scanner{
		Tokens: ms.Tokens,
	}

	tok := sc.Next()

	for tok != nil {
		if m, ok := tok.(*Macro); ok {
			if m.Name != "CHAPTER" && m.Name != "CHAPTER_TITLE" {
				tok = sc.Next()
				continue
			}

			ch := Chapter{
				Manuscript: ms,
			}

			var start, end int

			if m.Name == "CHAPTER" {
				count++

				ch.Count = count
				start = sc.Pos - 1

				tok = sc.Next()

				// Parse the rest of the chapter, which would include the
				// CHAPTER_TITLE macro if also specified. We don't want to
				// mistakenly count a chapter twice.
				goto parseChapter
			}

			if m.Name == "CHAPTER_TITLE" {
				count++

				ch.Count = count
				start = sc.Pos - 1

				tok = sc.Next()
			}

		parseChapter:
			for {
				if tok == nil {
					end = sc.Pos
					break
				}

				if m, ok := tok.(*Macro); ok {
					if m.Name == "COLLATE" {
						end = sc.Pos
						tok = sc.Next()
						break
					}
				}
				tok = sc.Next()
			}

			// Filter out the chapters that have been specified, if any. First we check
			// for chapter counts, then fallback to checking chapter titles.
			if len(set) > 0 {
				n := strconv.Itoa(count)

				if _, ok := set[n]; !ok {
					title := ch.Title()

					if _, ok := set[title]; !ok {
						continue
					}
				}
			}

			ch.Tokens = make([]Token, end-start)
			copy(ch.Tokens, ms.Tokens[start:end])

			chapters = append(chapters, &ch)

			if m, ok := tok.(*Macro); ok {
				// If the next token we encounter is for a macro, then immediately
				// go to the next iteration to parse it, so we don't skip it.
				if m.Name == "CHAPTER" || m.Name == "CHAPTER_TITLE" {
					continue
				}
			}
		}
		tok = sc.Next()
	}
	return chapters, nil
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
