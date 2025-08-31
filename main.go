package main

import (
	"fmt"
	"os"
)

func run(args []string) error {
	cmds := CommandSet{
		Argv0: os.Args[0],
		Long: `book is a tool for working with groff manuscripts for writing books.

Usage:

    book <command> [arguments]
`,
	}

	cmds.Add("cat", CatCmd)
	cmds.Add("clean", CleanCmd)
	cmds.Add("ls", LsCmd)
	cmds.Add("new", NewCmd)
	cmds.Add("pub", PubCmd)
	cmds.Add("wc", WcCmd)

	cmds.Add("help", HelpCmd(&cmds))

	return cmds.Parse(args[1:])
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
