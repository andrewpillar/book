package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

type Buffer struct {
	buf  []byte
	tmp  []rune
	pos  int
	eof  int
	line int
	col  int
}

func BufferString(s string) *Buffer {
	return &Buffer{
		buf: []byte(s),
		eof: len(s),
	}
}

func BufferFile(name string) (*Buffer, error) {
	b, err := os.ReadFile(name)

	if err != nil {
		return nil, err
	}

	return &Buffer{
		buf: b,
		eof: len(b),
	}, nil
}

func (b *Buffer) Seek(i int) {
	b.pos = i

	if b.pos < 0 {
		b.pos = 0
	}

	if b.pos > b.eof {
		b.pos = b.eof
	}
}

func (b *Buffer) Read(p []byte) (int, error) {
	if b.pos >= b.eof {
		return 0, io.EOF
	}

	n := copy(p, b.buf[b.pos:])
	b.pos += n
	return n, nil
}

func (b *Buffer) Get() rune {
	if b.pos >= b.eof {
		return -1
	}

	r := rune(b.buf[b.pos])
	w := 1

	if r >= utf8.RuneSelf {
		r, w = utf8.DecodeRune(b.buf[b.pos:])
	}

	b.pos += w
	b.col += w

	if r == '\n' {
		b.line++
		b.col = 0
	}
	return r
}

func (b *Buffer) Line() (string, bool) {
	r := b.Get()

	if r == -1 {
		return "", false
	}

	b.tmp = b.tmp[0:0]

	for r != '\n' {
		b.tmp = append(b.tmp, r)
		r = b.Get()
	}
	return string(b.tmp), true
}

type Macro struct {
	Name string
	Args []string
}

func ParseMacro(s string) *Macro {
	var m Macro

	buf := BufferString(s)

	tmp := make([]rune, 0, len(s))

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
	}
	return &m
}

type Chapter struct {
	Buffer   *Buffer
	Number   int
	Title    string
	Subtitle string
	Start    int
	End      int
}

func (c *Chapter) Content() string {
	buf := make([]byte, c.End-c.Start)

	c.Buffer.Seek(c.Start)
	c.Buffer.Read(buf)

	return string(buf)
}

func (c *Chapter) Text() string {
	var buf bytes.Buffer

	c.Buffer.Seek(c.Start)

	for {
		line, ok := c.Buffer.Line()

		if !ok || line == ".COLLATE" {
			break
		}

		if line == "" {
			continue
		}

		if line[0] == '.' {
			m := ParseMacro(line)

			// Special case, we want to include the character in the output too.
			if m.Name == "DROPCAP" {
				buf.WriteString(m.Args[0])
			}

			if m.Name == "PP" {
				buf.WriteString("\n")
			}
			continue
		}

		lbuf := BufferString(line)
		tmp := make([]byte, 0, len(line))

		r := lbuf.Get()

		// Strip the line of any inline formatting macros.
		for r != -1 {
			if r == '\\' {
				r = lbuf.Get()

				if r == '*' || r == '[' {
					for r != ']' {
						r = lbuf.Get()
					}
					r = lbuf.Get()
				}
			}

			tmp = utf8.AppendRune(tmp, r)
			r = lbuf.Get()
		}

		tmp = append(tmp, '\n')
		buf.Write(tmp)
		tmp = tmp[0:0]
	}
	return buf.String()
}

func (c *Chapter) Epigraph() []string {
	lines := make([]string, 0)

	c.Buffer.Seek(c.Start)

	for {
		line, ok := c.Buffer.Line()

		if !ok || line == ".COLLATE" {
			return nil
		}

		if line == ".EPIGRAPH" {
			for {
				line, ok = c.Buffer.Line()

				if !ok || line == ".EPIGRAPH OFF" {
					break
				}
				lines = append(lines, line)
			}
			break
		}
	}
	return lines
}

func (c *Chapter) Paragraphs() []string {
	paragraphs := make([]string, 0)

	c.Buffer.Seek(c.Start)

	var buf bytes.Buffer

	tmp := make([]byte, 0)

	for {
		line, ok := c.Buffer.Line()

		if !ok || line == ".COLLATE" {
			if buf.Len() > 0 {
				unfolded := strings.Replace(buf.String(), "\n", " ", -1)

				paragraphs = append(paragraphs, strings.TrimSuffix(unfolded, " "))
				buf.Reset()
			}
			break
		}

		if line == "" {
			continue
		}

		if line[0] == '.' {
			m := ParseMacro(line)

			if m.Name == "DROPCAP" {
				buf.WriteString(m.Args[0])
			}

			if m.Name == "EPIGRAPH" {
				line, ok = c.Buffer.Line()

				if !ok || line == ".COLLATE" {
					break
				}

				for {
					line, ok = c.Buffer.Line()

					if !ok {
						break
					}

					if line != "" {
						if line == ".EPIGRAPH OFF" {
							break
						}
					}
				}
			}

			if m.Name == "PP" {
				if buf.Len() > 0 {
					unfolded := strings.Replace(buf.String(), "\n", " ", -1)

					paragraphs = append(paragraphs, strings.TrimSuffix(unfolded, " "))
					buf.Reset()
				}
			}
			continue
		}

		lbuf := BufferString(line)

		r := lbuf.Get()

		// Strip the line of any inline formatting macros.
		for r != -1 {
			if r == '\\' {
				r = lbuf.Get()

				if r == '*' || r == '[' {
					for r != ']' {
						r = lbuf.Get()
					}
					r = lbuf.Get()
				}
			}

			tmp = utf8.AppendRune(tmp, r)
			r = lbuf.Get()
		}

		tmp = append(tmp, '\n')
		buf.Write(tmp)
		tmp = tmp[0:0]
	}
	return paragraphs
}

func (c *Chapter) WordCount() int {
	lines := strings.Split(c.Text(), "\n")
	wc := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		wc += len(strings.Split(strings.TrimSpace(line), " "))
	}
	return wc
}

type Manuscript struct {
	Buffer     *Buffer
	PrintStyle string
	Title      string
	Subtitle   string
	Author     string
	Chapters   []*Chapter
}

func LoadManuscript(name string) (*Manuscript, error) {
	buf, err := BufferFile(name)

	if err != nil {
		return nil, err
	}

	ms := Manuscript{
		Buffer: buf,
	}

	line, ok := ms.Buffer.Line()

	// Number of the chapter.
	num := 0

	for ok {
		if line != "" {
			if line[0] == '.' {
				m := ParseMacro(line)

				switch m.Name {
				case "DOCTITLE":
					ms.Title = m.Args[0]
				case "PRINTSTYLE":
					ms.PrintStyle = m.Args[0]
				case "SUBTITLE":
					ms.Subtitle = m.Args[0]
				case "AUTHOR":
					ms.Author = m.Args[0]
				case "CHAPTER_TITLE":
					num++

					start := ms.Buffer.pos

					line, ok = ms.Buffer.Line()

					for {
						if !ok || line == ".COLLATE" {
							break
						}
						line, ok = ms.Buffer.Line()
					}

					end := ms.Buffer.pos

					ms.Chapters = append(ms.Chapters, &Chapter{
						Buffer: ms.Buffer,
						Number: num,
						Title:  m.Args[0],
						Start:  start,
						End:    end,
					})
				}
			}
		}
		line, ok = ms.Buffer.Line()
	}

	ms.Buffer.Seek(0)

	return &ms, nil
}

func (m *Manuscript) WordCount() int {
	wc := 0

	for _, ch := range m.Chapters {
		wc += ch.WordCount()
	}
	return wc
}
