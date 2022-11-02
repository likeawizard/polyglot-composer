# polyglot-composer
A library to compose polyglot opening books from PGN file containing one or multiple games.

## Build
Run `make build` to compile polyglot-composer

## Usage
-o flag is optional output file name. Defaults to poly_out.bin

`polyglot-composer -pgn <pgn_input.pgn> [-o <book.bin>]`

## Known issues and planned features
* Annotated PGNs currently not supported
* Add filtering on PGN tags (ELO ranges and differences, Time Control, Variant, etc...)
* Makes no distinction between games won by checkmate or timeout or other termination of games
* Compose book directly from lichess user id
* Compose FEN list for texel tuning based from PGNs and book

## Acknowledgments
 * H.G. Muller for providing the polyglot book format: [polyglot book format definition](http://hgm.nubati.net/book_format.html)