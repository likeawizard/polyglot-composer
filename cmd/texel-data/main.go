package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/likeawizard/polyglot-composer/pkg/pgn"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	var pgnPath, outPath string
	flag.StringVar(&pgnPath, "pgn", "", "PGN path")
	flag.StringVar(&outPath, "o", "texel_data.txt", "Texel data output")
	flag.Parse()

	if pgnPath == "" {
		fmt.Println("no pgn provided")
		return
	}
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
		pp, err := pgn.NewPGNParser(path, false)
		if err != nil {
			fmt.Printf("could not load pgn file: %s with error: %s\n", path, err)
			continue
		}

		fenChan := make(chan string)
		pgnChan := make(chan *pgn.PGN, 20)
		go func() {
			for pp.Scan(ctx) {
				pgnChan <- pp.PGN()
			}
			pp.Close()
			pp.Progress(true)
			close(pgnChan)
		}()

		var writeWG sync.WaitGroup
		writeWG.Add(1)
		go func() {
			file, err := os.Create(outPath)
			if err != nil {
				fmt.Println("failed opening file for writing: ", path)
			}
			defer file.Close()

			writer := bufio.NewWriter(file)
			for fen := range fenChan {
				_, _ = writer.WriteString(fen)
			}
			writer.Flush()
			writeWG.Done()
		}()

		var wg sync.WaitGroup
		for game := range pgnChan {
			wg.Add(1)
			go func(game *pgn.PGN) {
				fens := game.GetFENs()
				for i := range fens {
					fenChan <- fens[i]
				}
				defer wg.Done()
			}(game)
		}

		fmt.Println()
		wg.Wait()
		close(fenChan)
		writeWG.Wait()
	}
	fmt.Printf("Book saved: %v\n", outPath)
}
