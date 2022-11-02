package main

import (
	"flag"
	"fmt"

	"github.com/likeawizard/polyglot-composer/pkg/pgn"
	"github.com/likeawizard/polyglot-composer/pkg/polyglot"
)

func main() {
	var pgn_path, out_path string
	flag.StringVar(&pgn_path, "pgn", "", "PGN path")
	flag.StringVar(&out_path, "o", "poly_out.bin", "Polyglot book output name. Default: poly_out.bin")
	flag.Parse()

	pgns := pgn.ParsePGN("games.pgn")
	if pgn_path == "" || len(pgns) == 0 {
		fmt.Println("could not load pgn")
		return
	}
	pb := polyglot.BuildFromPGNs(pgns)
	pb.SaveBook(out_path)
}
