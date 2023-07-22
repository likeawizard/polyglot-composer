package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/likeawizard/polyglot-composer/pkg/pgn"
	"github.com/likeawizard/polyglot-composer/pkg/polyglot"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	var pgnPath, outPath string
	flag.StringVar(&pgnPath, "pgn", "", "PGN path")
	flag.StringVar(&outPath, "o", "poly_out.bin", "Polyglot book output name.")
	flag.IntVar(&polyglot.MoveLimit, "d", 40, "Move depth limit.")
	flag.Parse()

	if pgnPath == "" {
		fmt.Println("no pgn provided")
		return
	}
	pb := polyglot.NewPolyglotBook()
	sources, err := pgn.ParsePath(pgnPath)
	if err != nil {
		fmt.Printf("could not parse pgn path: %s", err)
	}

SourceLoop:
	for _, path := range sources {
		select {
		case <-ctx.Done():
			break SourceLoop
		default:
		}
		pp, err := pgn.NewPGNParser(path, true)

		if err != nil {
			fmt.Printf("could not load pgn file: %s with error: %s\n", path, err)
			continue
		}

		pgnChan := make(chan *pgn.PGN, 20)
		go func() {
			for pp.Scan(ctx) {
				pgnChan <- pp.PGN()
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

	pb.SaveBook(outPath)
	fmt.Printf("Book saved: %v\n", outPath)
}
