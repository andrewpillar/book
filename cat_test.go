package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCat(t *testing.T) {
	path := filepath.Join("testdata", "chapters.mom")

	buf := CaptureOutput(CatCmd)

	tests := []struct {
		args []string
		want string
	}{
		{
			[]string{path},
			`CHAPTERS EXAMPLE by Andrew Pillar

CHAPTER I
THE FIRST

The first.

CHAPTER II
THE SECOND

The second.

CHAPTER III
THE THIRD

The third.`,
		},
		{
			[]string{path, "2"},
			`CHAPTERS EXAMPLE by Andrew Pillar

CHAPTER II
THE SECOND

The second.`,
		},
		{
			[]string{path, "2:3"},
			`CHAPTERS EXAMPLE by Andrew Pillar

CHAPTER II
THE SECOND

The second.

CHAPTER III
THE THIRD

The third.`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			if err := catCmd(CatCmd, test.args); err != nil {
				t.Fatalf("catCmd(CatCmd, %v): %v\n", test, err)
			}

			if diff := cmp.Diff(test.want, buf.String()); diff != "" {
				t.Fatalf("catCmd(CatCmd, %v) mismatch (-want +got):\n%s", test.args, diff)
			}
			buf.Reset()
		})
	}
}
