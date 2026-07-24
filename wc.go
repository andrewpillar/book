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

type ChapterNotFoundError string

func (e ChapterNotFoundError) Error() string {
	return fmt.Sprintf("no such chapter %s", string(e))
}

var ErrNoChapters = errors.New("no chapters")

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

	ms, err := ParseManuscript(file)

	if err != nil {
		return err
	}

	chapters, err := ms.Chapters()

	if err != nil {
		return err
	}

	if len(args) > 1 {
		if len(chapters) == 0 {
			return ErrNoChapters
		}

		chapter := args[1]

		for i, ch := range chapters {
			n, err := strconv.Atoi(chapter)

			if err == nil {
				if i+1 == n {
					cmd.Println(formatNumber(ch.WordCount()))
					return nil
				}
			}

			if ch.Title() == chapter {
				cmd.Println(formatNumber(ch.WordCount()))
				return nil
			}
		}
		return ChapterNotFoundError(chapter)
	}

	wc := ms.WordCount()

	if len(chapters) > 0 {
		cmd.Println("Average chapter word count:", formatNumber(wc/len(chapters)))
		cmd.Println("Manuscript word count:     ", formatNumber(wc))
		return nil
	}

	cmd.Println("Manuscript word count:", formatNumber(wc))
	return nil
}
