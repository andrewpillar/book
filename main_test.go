package main

import (
	"testing"
)

func Test_Cat(t *testing.T) {
	if err := run([]string{"book", "cat", "testdata/dracula.mom"}); err != nil {
		t.Fatal(err)
	}
}

func Test_Ls(t *testing.T) {
	if err := run([]string{"book", "ls", "testdata/dracula.mom"}); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"book", "ls", "-wc", "testdata/dracula.mom"}); err != nil {
		t.Fatal(err)
	}
}

func Test_Wc(t *testing.T) {
	if err := run([]string{"book", "wc", "testdata/dracula.mom"}); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"book", "wc", "testdata/dracula.mom", "CHAPTER ONE"}); err != nil {
		t.Fatal(err)
	}
}
