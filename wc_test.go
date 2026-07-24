package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWc(t *testing.T) {
	path := filepath.Join("testdata", "chapters.mom")

	buf := CaptureOutput(CatCmd)

	tests := []struct {
		args []string
		want string
		err  error
	}{
		{
			[]string{path},
			`Average chapter word count: 2
Manuscript word count:      6
`,
			nil,
		},
		{
			[]string{path, "2"},
			"2\n",
			nil,
		},
		{
			[]string{path, "THE SECOND"},
			"2\n",
			nil,
		},
		{
			[]string{path, "foo"},
			"",
			ChapterNotFoundError("foo"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			err := wcCmd(CatCmd, test.args)

			if err != test.err {
				t.Fatalf("wcCmd(CatCmd, %v): %v\n", test, err)
			}

			if diff := cmp.Diff(test.want, buf.String()); diff != "" {
				t.Fatalf("wcCmd(CatCmd, %v) mismatch (-want +got):\n%s", test.args, diff)
			}
			buf.Reset()
		})
	}
}
