package main

import (
	"bytes"
	"fmt"
	"io"
)

func fprint(w io.Writer) func(...any) (int, error) {
	return func(a ...any) (int, error) {
		return fmt.Fprint(w, a...)
	}
}

func fprintf(w io.Writer) func(string, ...any) (int, error) {
	return func(format string, a ...any) (int, error) {
		return fmt.Fprintf(w, format, a...)
	}
}

func fprintln(w io.Writer) func(...any) (int, error) {
	return func(a ...any) (int, error) {
		return fmt.Fprintln(w, a...)
	}
}

func CaptureOutput(cmd *Command) *bytes.Buffer {
	var buf bytes.Buffer

	cmd.Print = fprint(&buf)
	cmd.Printf = fprintf(&buf)
	cmd.Println = fprintln(&buf)

	return &buf
}
