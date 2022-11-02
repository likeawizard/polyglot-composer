# polyglot-composer
A library to compose polyglot opening books from PGN file containing one or multiple games. Supports bz2 compressed PGNs

## Features
* Multi file support - set a list of files to parse into a single book
* Supports large files. Polyglot weights are limited by uint16 (65535). This means if a move is encountered that many times (reasonable for 1. e4 ..., 1. d4 ...) the entries with excessive weights are normalized.
* Normalization can cause low weight moves to be dropped entirely
* Supports both annotated and raw PGN. `Event` tag is required and must be the first tag in the tag list as it currently works as a separator for games.

## Build
Run `make build` to compile polyglot-composer

## Usage
-o flag is optional output file name. Defaults to poly_out.bin
-pgn single or comma delimited files

`polyglot-composer -pgn <pgn_input.pgn>|<pgn1.pgn,pgn2.pgn.bz2,...> [-o <book.bin>]`

## Known issues and planned features
* ~~Annotated PGNs currently not supported~~ Supported.
* Allow a directory to be passed as input and parse all files within
* Add filtering on PGN tags (ELO ranges and differences, Time Control, Variant, etc...)
* Makes no distinction between games won by checkmate or timeout or other termination of games
* Compose book directly from lichess user id
* Compose FEN list for texel tuning based from PGNs and book

## Acknowledgments
 * H.G. Muller for providing the polyglot book format: [polyglot book format definition](http://hgm.nubati.net/book_format.html)