package main

var CatCmd = &Command{
	Usage: "cat <file> [chapter,...]",
	Short: "print out text content of the manuscript",
	Run:   catCmd,
}

func catCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	file := args[0]
	args = args[1:]

	ms, err := ParseManuscript(file)

	if err != nil {
		return err
	}

	sc := Scanner{
		Tokens: ms.Tokens,
	}

	if len(args) > 0 {
		chapters, err := ms.Chapters(args...)

		if err != nil {
			return err
		}

		sc.Tokens = sc.Tokens[0:0]

		for _, tok := range ms.Tokens {
			if m, ok := tok.(*Macro); ok {
				switch m.Name {
				case "DOCTITLE":
					sc.Tokens = append(sc.Tokens, m)
				case "AUTHOR":
					sc.Tokens = append(sc.Tokens, m)
				}
			}
		}

		for _, ch := range chapters {
			sc.Tokens = append(sc.Tokens, ch.Tokens...)
		}

		last := sc.Tokens[len(sc.Tokens)-1]

		if m, ok := last.(*Macro); ok {
			// Remove trailing COLLATE macro to prevent superfluous newlines
			// from being printed.
			if m.Name == "COLLATE" {
				sc.Tokens = sc.Tokens[:len(sc.Tokens)-1]
			}
		}
	}

	tok := sc.Next()

	for tok != nil {
		if m, ok := tok.(*Macro); ok {
			switch m.Name {
			case "DOCTITLE":
				title := m.Arg(0)
				author := ""

				tok = sc.Next()

				for {
					if tok == nil {
						return nil
					}
					if m, ok := tok.(*Macro); ok && m.Name == "AUTHOR" {
						author = m.Arg(0)
						break
					}
					tok = sc.Next()
				}

				cmd.Println(title, "by", author)
				cmd.Println()
			case "CHAPTER":
				chapterStr := "CHAPTER"

				if tok := sc.Peek(); tok != nil {
					if m, ok := tok.(*Macro); ok && m.Name == "CHAPTER_STRING" {
						chapterStr = m.Arg(0)
						sc.Next()
					}
				}
				cmd.Println(chapterStr, m.Arg(0))
			case "CHAPTER_TITLE":
				cmd.Println(m.Arg(0))
				cmd.Println()
			case "EPIGRAPH":
				cmd.Println()
				PrintEpigraph(cmd, &sc)
			case "PP":
				PrintParagraph(cmd, &sc)

				if sc.Peek() != nil {
					cmd.Println()
					cmd.Println()
				}
			case "COLLATE":
			}
		}
		tok = sc.Next()
	}
	return nil
}
