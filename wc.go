package main

import (
	"errors"
	"fmt"
	"strconv"
)

var WcCmd = &Command{
	Usage: "wc <file> [chapter]",
	Short: "display manuscript word count and average chapter word count",
	Run:   wcCmd,
}

var ErrNoSuchChapter = errors.New("no such chapter")

func formatNumber(n int) string {
	s := strconv.FormatInt(int64(n), 10)

	for i := len(s); i > 3; {
		i -= 3
		s = s[:i] + "," + s[i:]
	}
	return s
}

func wcCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	file := args[0]
	chapter := ""

	if len(args) > 1 {
		chapter = args[1]
	}

	ms, err := ParseManuscript(file)

	if err != nil {
		return err
	}

	chapters := ms.Chapters()

	if chapter != "" {
		for _, ch := range chapters {
			if ch.Title() == chapter {
				fmt.Println(formatNumber(ch.WordCount()))
				return nil
			}
		}
	}

	sum := 0

	for _, ch := range chapters {
		sum += ch.WordCount()
	}

	fmt.Println("Average chapter word count:", formatNumber(sum/len(chapters)))
	fmt.Println("Manuscript word count:     ", formatNumber(sum))
	return nil
}
