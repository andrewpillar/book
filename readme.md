# book

* [Installation](#installation)
* [Usage](#usage)
* [Publishing](#publishing)

book is a simple tool for working with [groff mom][] manuscript files. It
provides word counts, either for the entire manuscript, or for each chapter,
and allows for publishing either parts of the manuscript or the entire
manuscript into either PDF or DOCX formats.

[groff mom]: https://www.schaffter.ca/mom/mom-01.html

# Installation

To install, simply clone the repository and run `make install`,

    $ git clone https://github.com/andrewpillar/book
    $ cd book/
    $ make install

This will install the binary into the `GOPATH`, ensure this is part of your
`PATH` so the binary can be invoked.

# Usage

book can be used for working with pre-existing manuscript files, or for
creating new manuscripts. New manuscripts can be created via the `book new`
command. This will create a new `.mom` file and open it up for editing via the
program specified via the `EDITOR` environment variable. Authorship information
is derived from the git configuration, it is assumed that book will be used
within a git repository.

    $ book new "Dracula"

The newly created `.mom` file will be populated with the given title, copyright
information, and the authorship taken from the git configuration,

    $ cat dracula.mom
    .DOCTITLE   "My Book"
    .PRINTSTYLE "TYPESET"
    .TITLE      DOC_COVER "\*[$DOCTITLE]"
    .AUTHOR     "Bram Stoker"
    .COPYRIGHT  DOC_COVER "1897 \*[$AUTHOR]"
    .DOC_COVER  TITLE AUTHOR COPYRIGHT
    
    .DOCTYPE CHAPTER
    .TITLE   "\*[$DOCTITLE]"

Information of the manuscript can be viewed via the `ls`, and `wc` commands,
which can be used for listing the chapters and providing a word count,
repectively.

    $ book ls dracula.mom
    CHAPTER ONE
    CHAPTER TWO

The `-n` flag can be given to provide the chapter number, and the `-wc` flag
can be given to provide the word count for the chapter,

    $ book ls -n -wc dracula.mom
    1 CHAPTER ONE  5,701
    2 CHAPTER TWO  5,475

Word count for the overall manuscript can be retrieved via `book wc`,

    $ book wc dracula.mom
    Average chapter word count: 5,588
    Manuscript word count:      11,176

A chapter name can be given to get the word count for an individual chapter,

    $ book wc dracula.mom "CHAPTER ONE"
    5,701

This repository includes an [example manuscript][] to demonstrate the groff mom
format which should be used as an addition to the [documentation][] of mom.

[example manuscript]: /testdata/dracula.mom

[documentation]: https://schaffter.ca/mom/momdoc/toc.html

# Publishing

Manuscripts can be published into both the PDF and DOCX formats, via the `pub`
command,

    $ book pub -f pdf dracula.mom

Individual chapters of a manuscript can be published via passing the chapter
names as additional arguments,

    $ book pub -f pdf dracula.mom "CHAPTER ONE"

## PDF

The PDF format requires [groff][]. If using Linux, then this should already be
installed, along with the mom macro set.

[groff]: https://www.gnu.org/software/groff/

To publish in the PDF format, simply specify it via the `-f` flag with the `pub`
command,

    $ book pub -f pdf dracula.mom
    dracula.pdf

This will produce the name of the published PDF file as the output upon success.

Under the hood this runs the following groff command,

    $ groff -k -mom -T pdf <file>.mom > <file>.pdf

## DOCX

Pretty much every literary agent expects manuscripts to be submitted in the DOCX
format. To publish in the DOCX format, simply specify it via the `-f` flag with
the `pub` command,

    $ book pub -f docx dracula.mom
    dracula.docx

### A note on DOCX

The manuscript produced in the DOCX format will not have formatting parity with
the PDF format. At the bare minimum it will ensure:

* Times New Roman
* Font size 12pt
* Double spaced lines

There are many features available via the groff mom macro set that are not
implemented in the DOCX format produced via book. Under the hood this uses the
[godocx][] library for producing DOCX files, which itself was inspired by the
[python-docx][] library.

[godocx]: https://pkg.go.dev/github.com/gomutex/godocx

[python-docx]: https://python-docx.readthedocs.io/en/latest/
