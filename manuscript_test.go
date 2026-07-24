package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestManuscript(t *testing.T) {
	file := filepath.Join("testdata", "dracula.mom")

	ms, err := ParseManuscript(file)

	if err != nil {
		t.Fatalf("ParseManuscript(%q): %v\n", file, err)
	}

	if title := ms.DocTitle(); title != "DRACULA" {
		t.Fatalf("ms.DocTitle() = %q, want = %q\n", title, "DRACULA")
	}

	if style := ms.PrintStyle(); style != "TYPESET" {
		t.Fatalf("ms.PrintStyle() = %q, want = %q\n", style, "TYPESET")
	}

	m := ms.Macro("COPYRIGHT")

	if val := m.Arg(1); val != `1897 \*[$AUTHOR]` {
		t.Fatalf("m.Arg(1) = %q, want = %q\n", val, `1897 \*[$AUTHOR]`)
	}

	if author := ms.Author(); author != "Bram Stoker" {
		t.Fatalf("ms.Author() = %q, want = %q\n", author, "Bram Stoker")
	}

	chapters, err := ms.Chapters()

	if err != nil {
		t.Fatalf("ms.Chapters(): %v\n", err)
	}

	if l := len(chapters); l != 2 {
		t.Fatalf("len(chapters) = %v, want = %v\n", l, 2)
	}

	b, err := os.ReadFile(file)

	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v\n", file, err)
	}

	want := sha256.New()
	want.Write(b)

	got := sha256.New()
	ms.WriteTo(got)

	if diff := cmp.Diff(want.Sum(nil), got.Sum(nil)); diff != "" {
		t.Fatalf("ms.WriteTo() mismatch (-want +got):\n%s", diff)
	}
}

func TestChapters(t *testing.T) {
	file := filepath.Join("testdata", "chapters.mom")

	ms, err := ParseManuscript(file)

	if err != nil {
		t.Fatalf("ParseManuscript(%q): %v\n", file, err)
	}

	tests := []struct {
		names []string
		want  []*Chapter
	}{
		{
			nil,
			[]*Chapter{
				{
					Manuscript: ms,
					Count:      1,
					Tokens: []Token{
						&Macro{
							Raw:  []rune(".CHAPTER I"),
							Name: "CHAPTER",
							Args: []string{"I"},
						},
						&Macro{
							Raw:  []rune(`.CHAPTER_TITLE "THE FIRST"`),
							Name: "CHAPTER_TITLE",
							Args: []string{"THE FIRST"},
						},
						&Macro{
							Raw:  []rune(".START"),
							Name: "START",
						},
						&Macro{
							Raw:  []rune(".PP"),
							Name: "PP",
						},
						&Text{
							Value: "The first.",
						},
						&Macro{
							Raw:  []rune(".COLLATE"),
							Name: "COLLATE",
						},
					},
				},
				{
					Manuscript: ms,
					Count:      2,
					Tokens: []Token{
						&Macro{
							Raw:  []rune(".CHAPTER II"),
							Name: "CHAPTER",
							Args: []string{"II"},
						},
						&Macro{
							Raw:  []rune(`.CHAPTER_TITLE "THE SECOND"`),
							Name: "CHAPTER_TITLE",
							Args: []string{"THE SECOND"},
						},
						&Macro{
							Raw:  []rune(".START"),
							Name: "START",
						},
						&Macro{
							Raw:  []rune(".PP"),
							Name: "PP",
						},
						&Text{
							Value: "The second.",
						},
						&Macro{
							Raw:  []rune(".COLLATE"),
							Name: "COLLATE",
						},
					},
				},
				{
					Manuscript: ms,
					Count:      3,
					Tokens: []Token{
						&Macro{
							Raw:  []rune(".CHAPTER III"),
							Name: "CHAPTER",
							Args: []string{"III"},
						},
						&Macro{
							Raw:  []rune(`.CHAPTER_TITLE "THE THIRD"`),
							Name: "CHAPTER_TITLE",
							Args: []string{"THE THIRD"},
						},
						&Macro{
							Raw:  []rune(".START"),
							Name: "START",
						},
						&Macro{
							Raw:  []rune(".PP"),
							Name: "PP",
						},
						&Text{
							Value: "The third.",
						},
					},
				},
			},
		},
		{
			[]string{"2"},
			[]*Chapter{
				{
					Manuscript: ms,
					Count:      2,
					Tokens: []Token{
						&Macro{
							Raw:  []rune(".CHAPTER II"),
							Name: "CHAPTER",
							Args: []string{"II"},
						},
						&Macro{
							Raw:  []rune(`.CHAPTER_TITLE "THE SECOND"`),
							Name: "CHAPTER_TITLE",
							Args: []string{"THE SECOND"},
						},
						&Macro{
							Raw:  []rune(".START"),
							Name: "START",
						},
						&Macro{
							Raw:  []rune(".PP"),
							Name: "PP",
						},
						&Text{
							Value: "The second.",
						},
						&Macro{
							Raw:  []rune(".COLLATE"),
							Name: "COLLATE",
						},
					},
				},
			},
		},
		{
			[]string{"2:3"},
			[]*Chapter{
				{
					Manuscript: ms,
					Count:      2,
					Tokens: []Token{
						&Macro{
							Raw:  []rune(".CHAPTER II"),
							Name: "CHAPTER",
							Args: []string{"II"},
						},
						&Macro{
							Raw:  []rune(`.CHAPTER_TITLE "THE SECOND"`),
							Name: "CHAPTER_TITLE",
							Args: []string{"THE SECOND"},
						},
						&Macro{
							Raw:  []rune(".START"),
							Name: "START",
						},
						&Macro{
							Raw:  []rune(".PP"),
							Name: "PP",
						},
						&Text{
							Value: "The second.",
						},
						&Macro{
							Raw:  []rune(".COLLATE"),
							Name: "COLLATE",
						},
					},
				},
				{
					Manuscript: ms,
					Count:      3,
					Tokens: []Token{
						&Macro{
							Raw:  []rune(".CHAPTER III"),
							Name: "CHAPTER",
							Args: []string{"III"},
						},
						&Macro{
							Raw:  []rune(`.CHAPTER_TITLE "THE THIRD"`),
							Name: "CHAPTER_TITLE",
							Args: []string{"THE THIRD"},
						},
						&Macro{
							Raw:  []rune(".START"),
							Name: "START",
						},
						&Macro{
							Raw:  []rune(".PP"),
							Name: "PP",
						},
						&Text{
							Value: "The third.",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.names), func(t *testing.T) {
			chapters, err := ms.Chapters(test.names...)

			if err != nil {
				t.Fatalf("ms.Chapters(): %v\n", err)
			}

			if diff := cmp.Diff(test.want, chapters); diff != "" {
				t.Fatalf("chapters mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want []Token
	}{
		{
			"italicised inline escape",
			`\*[IT]italicised\*[PREV]`,
			[]Token{
				&Inline{
					Escape: "IT",
				},
				&Text{
					Value: "italicised",
				},
				&Inline{
					Escape: "PREV",
				},
			},
		},
		{
			"bold italicised inline escape",
			`\*[BDI]bold italicised\*[PREV]`,
			[]Token{
				&Inline{
					Escape: "BDI",
				},
				&Text{
					Value: "bold italicised",
				},
				&Inline{
					Escape: "PREV",
				},
			},
		},
		{
			"unclosed inline escape",
			`\*[BDI]bold italicised`,
			[]Token{
				&Inline{
					Escape: "BDI",
				},
				&Text{
					Value: "bold italicised",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			toks := Tokenize(test.str)

			if diff := cmp.Diff(test.want, toks); diff != "" {
				t.Errorf("Tokenize(%q) mismatch (-want +got):\n%s", test.str, diff)
			}
		})
	}
}
