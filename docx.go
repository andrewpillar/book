package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

const HeadingSize = 36

func BuildCover(doc domain.Document, ms *Manuscript) error {
	for i := 0; i < 10; i++ {
		if _, err := doc.AddParagraph(); err != nil {
			return err
		}
	}

	for i, txt := range []string{ms.DocTitle(), "by", ms.Author()} {
		p, err := doc.AddParagraph()

		if err != nil {
			return err
		}

		if err := p.SetAlignment(domain.AlignmentCenter); err != nil {
			return err
		}

		r, err := p.AddRun()

		if err != nil {
			return err
		}

		if err := r.SetText(txt); err != nil {
			return err
		}

		if err := r.SetSize(HeadingSize); err != nil {
			return err
		}

		if i == 0 {
			if err := r.SetBold(true); err != nil {
				return err
			}
			continue
		}

		if err := r.SetItalic(true); err != nil {
			return err
		}
	}

	s, err := doc.DefaultSection()

	if err != nil {
		return err
	}

	ftr, err := s.Footer(domain.FooterDefault)

	if err != nil {
		return err
	}

	p, err := ftr.AddParagraph()

	if err != nil {
		return err
	}

	p.SetAlignment(domain.AlignmentRight)

	r, err := p.AddRun()

	if err != nil {
		return err
	}

	r.SetText(fmt.Sprintf("© %d %s", time.Now().Year(), ms.Author()))
	return nil
}

func BuildText(p domain.Paragraph, txt string) error {
	sc := Scanner{
		Tokens: Tokenize(txt),
	}

	tok := sc.Next()

	for tok != nil {
		switch v := tok.(type) {
		case *Inline:
			r, err := p.AddRun()

			if err != nil {
				return err
			}

			switch v.Escape {
			case "IT", "BD", "BDI":
				switch v.Escape {
				case "IT":
					r.SetItalic(true)
				case "BD":
					r.SetBold(true)
				case "BDI":
					r.SetItalic(true)
					r.SetBold(true)
				}

				tok = sc.Next()

			inner:
				for tok != nil {
					switch v := tok.(type) {
					case *Inline:
						if v.Escape == "PREV" {
							break inner
						}
					case *Text:
						r.SetText(v.Value)
					}
					tok = sc.Next()
				}
			case "lq":
				r, err := p.AddRun()

				if err != nil {
					return err
				}
				r.SetText("“")
			case "rq":
				r, err := p.AddRun()

				if err != nil {
					return err
				}
				r.SetText("”")
			}
		case *Text:
			r, err := p.AddRun()

			if err != nil {
				return err
			}
			r.SetText(v.Value)
		}
		tok = sc.Next()
	}
	return nil
}

const (
	LineSpacing = 480
	ParaIndent  = 567
)

func WriteToDOCX(name string, ms *Manuscript) error {
	doc := docx.NewDocument()
	doc.SetMetadata(&domain.Metadata{
		Title:   ms.DocTitle(),
		Creator: ms.Author(),
		Created: time.Now().Format(time.RFC3339),
	})

	s, err := doc.DefaultSection()

	if err != nil {
		return err
	}

	s.SetPageSize(domain.PageSizeA4)

	type FontSetter interface {
		SetDefaultFont(string) error
	}

	if f, ok := doc.(FontSetter); ok {
		f.SetDefaultFont("Times New Roman")
	}

	type FontSize interface {
		SetDefaultFontSize(int) error
	}

	if f, ok := doc.(FontSize); ok {
		f.SetDefaultFontSize(24)
	}

	if err := BuildCover(doc, ms); err != nil {
		return err
	}

	s2, err := doc.AddSectionWithBreak(domain.SectionBreakTypeNextPage)

	if err != nil {
		return err
	}

	ftr, err := s2.Footer(domain.FooterDefault)

	if err != nil {
		return err
	}

	ftrPara, err := ftr.AddParagraph()

	if err != nil {
		return err
	}

	if err := ftrPara.SetAlignment(domain.AlignmentCenter); err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		r, err := ftrPara.AddRun()

		if err != nil {
			return err
		}

		switch i {
		case 0:
			err = r.AddField(docx.NewPageNumberField())
		case 1:
			err = r.AddText(" of ")
		case 2:
			err = r.AddField(docx.NewPageCountField())
		}

		if err != nil {
			return err
		}
	}

	sc := Scanner{
		Tokens: ms.Tokens,
	}

	var buf bytes.Buffer

	firstPara := true

	tok := sc.Next()

	for tok != nil {
		if m, ok := tok.(*Macro); ok {
			switch m.Name {
			case "CHAPTER":
				str := "CHAPTER"

				if tok := sc.Peek(); tok != nil {
					if m, ok := tok.(*Macro); ok && m.Name == "CHAPTER_STRING" {
						str = m.Arg(0)
						sc.Next()
					}
				}

				p, err := doc.AddParagraph()

				if err != nil {
					return err
				}

				p.SetLineSpacing(domain.LineSpacing{
					Value: LineSpacing,
				})

				if err := p.SetAlignment(domain.AlignmentCenter); err != nil {
					return err
				}

				r, err := p.AddRun()

				if err != nil {
					return err
				}

				if err := r.SetSize(HeadingSize); err != nil {
					return err
				}

				if err := r.SetText(fmt.Sprintf("%s %s", str, m.Arg(0))); err != nil {
					return err
				}

				if err := r.SetBold(true); err != nil {
					return err
				}
			case "CHAPTER_TITLE":
				p, err := doc.AddParagraph()

				if err != nil {
					return err
				}

				p.SetLineSpacing(domain.LineSpacing{
					Value: LineSpacing,
				})

				if err := p.SetAlignment(domain.AlignmentCenter); err != nil {
					return err
				}

				r, err := p.AddRun()

				if err != nil {
					return err
				}

				if err := r.SetSize(HeadingSize); err != nil {
					return err
				}

				if err := r.SetText(m.Arg(0)); err != nil {
					return err
				}

				if err := r.SetBold(true); err != nil {
					return err
				}

				if err := r.SetItalic(true); err != nil {
					return err
				}
			case "EPIGRAPH":
				tok = sc.Next()

			epiLoop:
				for tok != nil {
					switch v := tok.(type) {
					case *Macro:
						if v.Name == "EPIGRAPH" {
							break epiLoop
						}
					case *Text:
						if v.Value != "" {
							p, err := doc.AddParagraph()

							if err != nil {
								return err
							}

							p.SetLineSpacing(domain.LineSpacing{
								Value: LineSpacing,
							})

							if err := p.SetAlignment(domain.AlignmentCenter); err != nil {
								return err
							}

							if err := BuildText(p, v.Value); err != nil {
								return err
							}
						}
					}
					tok = sc.Next()
				}
			case "PP":
				if tok := sc.Peek(); tok != nil {
					if m, ok := tok.(*Macro); ok {
						switch m.Name {
						case "DROPCAP":
							buf.WriteString(m.Arg(0))
							sc.Next()
						case "PP":
							tok = sc.Next()
							continue
						}
					}
				}

				tok = sc.Next()

				p, err := doc.AddParagraph()

				if err != nil {
					return err
				}

				if !firstPara {
					p.SetIndent(domain.Indentation{
						FirstLine: ParaIndent,
					})
				}

				p.SetLineSpacing(domain.LineSpacing{
					Value: LineSpacing,
				})

			paraLoop:
				for tok != nil {
					switch v := tok.(type) {
					case *Macro:
						if v.Name == "PP" {
							firstPara = false
							sc.Back()
							break paraLoop
						}
						if v.Name == "COLLATE" {
							firstPara = true
							sc.Back()
							break paraLoop
						}
						if v.Name == "RIGHT" {
							p.SetAlignment(domain.AlignmentRight)
						}
					case *Text:
						buf.WriteString(v.Value)
						buf.WriteString(" ")
					}
					tok = sc.Next()
				}

				if err := BuildText(p, strings.TrimSuffix(buf.String(), " ")); err != nil {
					return err
				}
				buf.Reset()
			case "COLLATE":
				firstPara = true

				if err := doc.AddPageBreak(); err != nil {
					return err
				}
			}
		}
		tok = sc.Next()
	}

	if err := doc.SaveAs(name); err != nil {
		return err
	}
	return nil
}
