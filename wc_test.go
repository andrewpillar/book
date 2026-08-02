package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWc(t *testing.T) {
	path := filepath.Join("testdata", "chapters.mom")

	buf := CaptureOutput(WcCmd)

	tests := []struct {
		args []string
		want string
		err  error
	}{
		{
			[]string{path},
			`Average chapter word count: 2
Manuscript word count:      8
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
		{
			[]string{
				filepath.Join("testdata", "no-chapters.mom"),
			},
			"Manuscript word count: 2\n",
			nil,
		},
		{
			[]string{
				filepath.Join("testdata", "no-chapters.mom"),
				"foo",
			},
			"",
			ErrNoChapters,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			err := wcCmd(WcCmd, test.args)

			if err != test.err {
				t.Fatalf("wcCmd(WcCmd, %v): %v\n", test, err)
			}

			if diff := cmp.Diff(test.want, buf.String()); diff != "" {
				t.Fatalf("wcCmd(WcCmd, %v) mismatch (-want +got):\n%s", test.args, diff)
			}
			buf.Reset()
		})
	}
}
