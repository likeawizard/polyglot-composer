package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/likeawizard/polyglot-composer/pkg/pgn"
	"github.com/likeawizard/polyglot-composer/pkg/polyglot"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	var pgn_path, out_path string
	flag.StringVar(&pgn_path, "pgn", "", "PGN path")
	flag.StringVar(&out_path, "o", "poly_out.bin", "Polyglot book output name. Default: poly_out.bin")
	flag.Parse()

	if pgn_path == "" {
		fmt.Println("no pgn provided")
		return
	}
	pb := polyglot.NewPolyglotBook()
	paths := strings.Split(pgn_path, ",")

	for _, path := range paths {
		pp, err := pgn.NewPGNParser(path)

		if err != nil {
			fmt.Printf("could not load pgn file: %s with error: %s\n", path, err)
			continue
		} else {
			fmt.Printf("Parsing '%s' ...\n", path)
		}

		pgnChan := make(chan *pgn.PGN, 20)
		go func() {
			for game := pp.Next(); game != nil; game = pp.Next() {
				pgnChan <- game
			}
			pp.Close()
			pp.Progress(true)
			close(pgnChan)
		}()

		var wg sync.WaitGroup
		for game := range pgnChan {
			wg.Add(1)
			go func(game *pgn.PGN) {
				defer wg.Done()
				pb.AddFromPGN(game)
			}(game)
		}
		fmt.Println()
		wg.Wait()
	}

	// fmt.Printf("games parsed: , book keys %d\n", len(pb.b))
	pb.SaveBook(out_path)
}
