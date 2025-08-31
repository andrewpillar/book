# book

book is a tool that provides a collection of utilities to make working with
manuscripts written in [groff mom][] easier. This will allow you to list the
chapters in a manuscript, publish individual chapters, or the entire manuscript
itself.

[groff mom]: https://schaffter.ca/mom/

## Installation

book requires that the [groff][] typesetter is installed as this is used under
the hood to actually handle the publishing of manuscripts. Recent releases
of groff should already have the mom macro set bundled with it.

[groff]: https://www.gnu.org/software/groff/

To install book, simply clone the repository and run `make install`. The binary
will be installed under the local `GOPATH`, ensure this is part of your `PATH`
before usage.

## Usage

The only thing book needs to be used is a pre-existing groff mom manuscript
file.


