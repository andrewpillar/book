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

    $ book new "DRACULA"

The newly created `.mom` file will be populated with the given title, copyright
information, and the authorship taken from the git configuration,

    $ cat dracula.mom
    .DOCTITLE   "DRACULA"
    .PRINTSTYLE "TYPESET"
    .AUTHOR     "Bram Stoker"
    .TITLE      DOC_COVER "\*[$DOCTITLE]"
    .PDF_TITLE  "\*[$AUTHOR] - \*[$DOCTITLE]"
    .COPYRIGHT  DOC_COVER "2026 \*[$AUTHOR]"
    .DOC_COVER  TITLE AUTHOR COPYRIGHT

    .RECTO_VERSO

    .HEADER_RULE OFF
    .HEADER_RECTO CENTER "\E*[$AUTHOR]"
    .HEADER_VERSO CENTER "\E*[$DOCTITLE]"


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

When no arguments are given to `ls`, then manuscripts in the current directory
will be listed,

    $ book ls
    DR. ACULA
    DRACULA 2: HARKER BITES BACK
    DRACULA

when given the `-wc` flag the word counts for each will be printed,

    $ book ls -wc
    DR. ACULA                       20,456
    DRACULA 2: HARKER BITES BACK    5,789
    DRACULA                         11,176

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

Chapter numbers can also be given in lieu of chapter titles,

    $ book pub -f pdf dracula.mom 1

The `pub` command will print out the name of the published manuscript as the
only output upon success. If chapters were specified then the name will be
formatted to reflect that. For example,

    $ book pub -f pdf dracula.mom 1 2

Chapter ranges can also be given by specifying the lower and upper bound between
a `:`. The below command would publish a PDF with only the first three chapters
of the manuscript,

    $ book pub -f pdf dracula.mom 1:3

Sometimes a literary agent will request a sample of a manuscript based off a
given word count. To only publish a certain number of words from the manuscript
pass the `-wc` flag,

    $ book pub -f docx -wc 10000 dracula.mom

This flag also works with specified chapters too,

    $ book pub -f docx -wc 7000 dracula.mom 1 2

## PDF

The PDF format requires [groff][] with the mom macro set. If using Linux or
MacOS, then this will already be installed.

[groff]: https://www.gnu.org/software/groff/

To publish in the PDF format, simply specify it via the `-f` flag with the `pub`
command,

    $ book pub -f pdf dracula.mom

Under the hood this runs the following pfdmom command,

    $ pdfmom -k <file>.mom > <file>.pdf

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
* Line indentations
* Page number and count in footer
* Copyright footer on cover page

There are many features available via the groff mom macro set that are not
implemented in the DOCX format produced via book.
