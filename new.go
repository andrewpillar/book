package main

import (
	"bytes"
	_ "embed"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//go:embed manuscript.tmpl
var manuscriptTmpl []byte

var (
	reSlug = regexp.MustCompile("[^a-zA-Z0-9]")
	reDupe = regexp.MustCompile("-{2,}")

	NewCmd = &Command{
		Usage: "new <title>",
		Short: "create a new manuscript",
		Run:   newCmd,
	}
)

func slug(s string) string {
	s = strings.TrimSpace(s)
	s = reSlug.ReplaceAllString(s, "-")
	s = reDupe.ReplaceAllString(s, "-")

	return strings.ToLower(strings.TrimPrefix(strings.TrimSuffix(s, "-"), "-"))
}

func GitUserName() (string, error) {
	var buf bytes.Buffer

	cmd := exec.Command("git", "config", "user.name")
	cmd.Stdin = os.Stdin
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String()[:buf.Len()-1], nil
}

func newCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	editor := os.Getenv("EDITOR")

	if editor == "" {
		return errors.New("EDITOR not set")
	}

	author, err := GitUserName()

	if err != nil {
		return err
	}

	name := slug(args[0])

	tmpl, err := template.New(name).Parse(string(manuscriptTmpl))

	if err != nil {
		return err
	}

	f, err := os.Create(name + ".mom")

	if err != nil {
		return err
	}

	defer f.Close()

	var data struct {
		Title      string
		PrintStyle string
		Author     string
		Year       int
	}

	data.Title = args[0]
	data.PrintStyle = "TYPESET"
	data.Author = author
	data.Year = time.Now().Year()

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	c := exec.Command(editor, f.Name())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = c.Stderr

	return c.Run()
}
