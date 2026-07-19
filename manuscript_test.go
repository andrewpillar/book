package main

import (
	"path/filepath"
	"testing"
)

func manuscriptFile() string {
	return filepath.Join("testdata", "dracula.mom")
}

func Test_ManuscriptDocTitle(t *testing.T) {
	ms, err := ParseManuscript(manuscriptFile())

	if err != nil {
		t.Fatal(err)
	}

	if title := ms.DocTitle(); title != "DRACULA" {
		t.Fatalf("unexpected DOCTITLE, expected=%q, got=%q\n", "DRACULA", title)
	}
}

func Test_ManuscriptAuthor(t *testing.T) {
	ms, err := ParseManuscript(manuscriptFile())

	if err != nil {
		t.Fatal(err)
	}

	if title := ms.Author(); title != "Bram Stoker" {
		t.Fatalf("unexpected AUTHOR, expected=%q, got=%q\n", "Bram Stoker", title)
	}
}

func Test_ManuscriptChapters(t *testing.T) {
	ms, err := ParseManuscript(manuscriptFile())

	if err != nil {
		t.Fatal(err)
	}

	chapters := ms.Chapters()

	if l := len(chapters); l != 2 {
		t.Fatalf("unexpected chapter count, expected=%d, got=%d\n", 2, l)
	}

	for _, ch := range chapters {
		t.Log(ch.Title())
	}
}
