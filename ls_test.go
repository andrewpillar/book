package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLs(t *testing.T) {
	t.Chdir("testdata")

	buf := CaptureOutput(LsCmd)

	tests := []struct {
		args []string
		want string
	}{
		{
			nil,
			`CHAPTERS EXAMPLE
DRACULA
NO CHAPTERS EXAMPLE
`,
		},
		{
			[]string{"-wc"},
			`CHAPTERS EXAMPLE         8
DRACULA             11,219
NO CHAPTERS EXAMPLE      2
`,
		},
		{
			[]string{"no-chapters.mom"},
			``,
		},
		{
			[]string{"dracula.mom"},
			`Chapter I
Chapter II
`,
		},
		{
			[]string{"-n", "chapters.mom"},
			`  1 THE FIRST
  2 THE SECOND
  3 THE THIRD
`,
		},
		{
			[]string{"-n", "-wc", "chapters.mom"},
			`  1 THE FIRST       4
  2 THE SECOND      2
  3 THE THIRD       2
`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			if err := lsCmd(LsCmd, test.args); err != nil {
				t.Fatalf("lsCmd(LsCmd, %v): %v\n", test, err)
			}

			if diff := cmp.Diff(test.want, buf.String()); diff != "" {
				t.Fatalf("lsCmd(LsCmd, %v) mismatch (-want +got):\n%s", test.args, diff)
			}
			buf.Reset()
		})
	}
}
