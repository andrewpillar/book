package main

import (
	"errors"
	"fmt"
)

var WcCmd = &Command{
	Usage: "wc <file> [chapter]",
	Short: "display manuscript word count and average chapter word count",
	Run:   wcCmd,
}

var ErrNoSuchChapter = errors.New("no such chapter")

func wcCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	file := args[0]
	chapter := ""

	if len(args) > 1 {
		chapter = args[1]
	}

	ms, err := LoadManuscript(file)

	if err != nil {
		return err
	}

	if chapter != "" {
		for _, ch := range ms.Chapters {
			if ch.Title == chapter {
				fmt.Println(formatNumber(ch.WordCount()))
				return nil
			}
		}
		return ErrNoSuchChapter
	}

	sum := 0

	for _, ch := range ms.Chapters {
		sum += ch.WordCount()
	}

	fmt.Println("Average chapter word count:", formatNumber(sum/len(ms.Chapters)))
	fmt.Println("Manuscript word count:     ", formatNumber(sum))

	return nil
}
