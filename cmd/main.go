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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	var interrupted bool

	var pgn_path, out_path string
	flag.StringVar(&pgn_path, "pgn", "", "PGN path")
	flag.StringVar(&out_path, "o", "poly_out.bin", "Polyglot book output name. Default: poly_out.bin")
	flag.IntVar(&polyglot.MoveLimit, "d", 40, "Move depth limit. Default: 40")
	flag.Parse()

	if pgn_path == "" {
		fmt.Println("no pgn provided")
		return
	}
	pb := polyglot.NewPolyglotBook()
	paths, err := pgn.PathParser(pgn_path)
	if err != nil {
		fmt.Printf("could not parse pgn path: %s", err)
	}

	for _, path := range paths {
		if interrupted {
			break
		}
		pp, err := pgn.NewPGNParser(path)

		if err != nil {
			fmt.Printf("could not load pgn file: %s with error: %s\n", path, err)
			continue
		} else {
			fmt.Printf("Parsing '%s' ...\n", path)
		}

		pgnChan := make(chan *pgn.PGN, 20)
		go func() {
			for pp.Scan() {
				select {
				case <-ctx.Done():
					cancel()
					interrupted = true
					break
				default:
					pgnChan <- pp.PGN()
				}
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

	pb.SaveBook(out_path)
}
