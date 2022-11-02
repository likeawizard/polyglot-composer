package main

import (
	"flag"
	"fmt"
	"strings"

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
	n := 0
	pb := make(polyglot.PolyglotBook, 0)
	paths := strings.Split(pgn_path, ",")

	for _, path := range paths {
		pp, err := pgn.NewPGNParser(path)

		if err != nil {
			fmt.Printf("could not load pgn file: %s with error: %s\n", path, err)
			continue
		} else {
			fmt.Printf("Parsing '%s' ...\n", path)
		}

		for pgn := pp.Next(); pgn != nil; pgn = pp.Next() {
			n++
			pb.AddFromPGN(pgn)
			if n%100 == 0 {
				pp.Progress(false)
			}
		}
		pp.Close()
		pp.Progress(true)
		fmt.Println()
	}

	fmt.Printf("games parsed:%d, book keys %d\n", n, len(pb))
	pb.SaveBook(out_path)
}
