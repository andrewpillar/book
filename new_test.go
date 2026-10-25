package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	os.Setenv("EDITOR", "true")

	args := []string{t.Name()}

	if err := newCmd(NewCmd, args); err != nil {
		t.Fatalf("newCmd(NewCmd, %v): %v\n", args, err)
	}

	path := "testnew.mom"

	t.Cleanup(func() {
		if !t.Failed() {
			os.Remove(path)
		}
	})

	b, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v\n", path, err)
	}

	author, err := GitUserName()

	if err != nil {
		t.Fatalf("GitUser(): %v\n", err)
	}

	want := fmt.Sprintf(`.ds STYLE TYPESET

.if '\*[STYLE]'TYPESET' .PAPER A5

.DOCTITLE   "%s"
.PRINTSTYLE "\*[STYLE]"
.AUTHOR     "%s"
.TITLE      DOC_COVER "\*[$DOCTITLE]"
.PDF_TITLE  "\*[$AUTHOR] - \*[$DOCTITLE]"
.COPYRIGHT  DOC_COVER "%d \*[$AUTHOR]"
.DOC_COVER  TITLE AUTHOR COPYRIGHT

.RECTO_VERSO

.HEADER_RULE OFF
.HEADER_RECTO CENTER "\E*[$AUTHOR]"
.HEADER_VERSO CENTER "\E*[$DOCTITLE]"

.if '\*[STYLE]'TYPESET' .T_MARGIN 1i
.if '\*[STYLE]'TYPESET' .B_MARGIN 1i
.if '\*[STYLE]'TYPESET' .L_MARGIN 0.75i
.if '\*[STYLE]'TYPESET' .R_MARGIN 0.75i

.if '\*[STYLE]'TYPESET' .PT_SIZE 10

`, t.Name(), author, time.Now().Year())

	if diff := cmp.Diff(want, string(b)); diff != "" {
		t.Fatalf("%q mismatch (-want +got):\n%s", path, diff)
	}
}
