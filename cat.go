package main

import (
	"fmt"
)

var CatCmd = &Command{
	Usage: "cat <file> [chapter]",
	Short: "display only the text contents of the manuscript",
	Run:   catCmd,
}

func catCmd(cmd *Command, args []string) error {
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

	shouldPrint := true

	for i, ch := range ms.Chapters {
		shouldPrint = chapter == "" || ch.Title == chapter

		if shouldPrint {
			if chapter == "" {
				fmt.Println(i+1, ch.Title)
			}
			fmt.Println(ch.Text())
		}
	}
	return nil
}
