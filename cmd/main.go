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

	if pgn_path == "" {
		fmt.Println("no pgn provided")
		return
	}

	pp, err := pgn.NewPGNParser(pgn_path)

	if err != nil {
		fmt.Println("could not load pgn")
		return
	} else {
		defer pp.Close()
	}

	n := 0
	pb := make(polyglot.PolyglotBook, 0)
	for pgn := pp.Next(); pgn != nil; pgn = pp.Next() {
		n++
		pb.AddFromPGN(pgn)
	}
	fmt.Printf("games parsed:%d, book keys %d\n", n, len(pb))
	pb.SaveBook(out_path)
}
