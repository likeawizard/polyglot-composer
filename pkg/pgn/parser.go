package pgn

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dsnet/compress/bzip2"

	"github.com/inhies/go-bytesize"
	"github.com/likeawizard/polyglot-composer/pkg/board"
)

func NewPGNParser(path string) (*PGNParser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening PGN: %s", err)
	}

	size := float64(1)
	stat, err := file.Stat()
	if err == nil {
		size = float64(stat.Size())
	}

	pp := &PGNParser{
		clock:      time.Now(),
		file:       file,
		totalBytes: bytesize.New(size)}

	if strings.HasSuffix(path, "bz2") {
		pp.isArchived = true
		bzReader, err := bzip2.NewReader(file, nil)
		if err != nil {
			return nil, fmt.Errorf("error opening bzip2: %s", err)
		}
		pp.bzipReader = bzReader
		pp.scanner = bufio.NewScanner(bufio.NewReader(bzReader))
	} else {
		pp.scanner = bufio.NewScanner(bufio.NewReader(file))
	}

	return pp, nil
}

func (pp *PGNParser) Scan() {
	pp.pgn = nil
	pgn := PGN{}
	var tag Tag
	var value string
	if pp.nextLine != "" {
		pgn.AddTag(parseTag(pp.nextLine))
		pp.nextLine = ""
	}
	for pp.scanner.Scan() {
		pp.lastBytes += bytesize.New(float64(len(pp.scanner.Bytes())))
		if time.Since(pp.clock) > time.Second {
			pp.Progress(false)
			pp.readBytes += pp.lastBytes
			pp.lastBytes = 0
			pp.clock = time.Now()
		}
		line := pp.scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if isTag(line) {
			tag, value = parseTag(line)
			if !pp.skipping {
				pp.skipping = !PreFilter(tag, value)
			}
			if tag == TAG_EVENT && pgn.Event != "" {
				if pp.skipping {
					pgn = PGN{Event: value}
					pp.skipping = false
				} else {
					pp.pgn = &pgn
					pp.nextLine = line
					return
				}
			}

			pgn.AddTag(tag, value)

		} else {
			pgn.Moves += line
		}
	}

	if pp.pgn == nil && pgn.Event != "" {
		pp.pgn = &pgn
	}

}

func (pp *PGNParser) Next() *PGN {
	pp.Scan()
	if pp.pgn != nil {
		pp.pgn.RemoveAnnotations()
	}
	pp.gameCount++
	return pp.pgn

}

func (pp *PGNParser) Progress(done bool) {
	ratio := 1.0
	if pp.bzipReader.InputOffset > 0 || pp.bzipReader.OutputOffset > 0 {
		ratio = float64(pp.bzipReader.OutputOffset) / float64(pp.bzipReader.InputOffset)
	}
	progress := math.Min(float64(pp.readBytes)/float64(pp.totalBytes)/ratio, 1)
	if done {
		progress = 1
	}
	barN := int(50 * progress)
	bar := "[" + strings.Repeat("#", barN) + strings.Repeat(".", 50-barN) + "]"
	output := fmt.Sprintf("%s %.2f%%, games: %d rate: %v/s read: %v, total: %v", bar, 100*progress, pp.gameCount, pp.lastBytes, pp.readBytes, pp.totalBytes*bytesize.New(ratio))
	if pp.isArchived {
		output += " (archived size)            "
	}

	fmt.Printf("%s\r", output)
}

func (pp *PGNParser) Close() error {
	return pp.file.Close()
}

// TODO: move limit
func (pgn *PGN) MovesToUCI() []string {
	const moveLimit = 40

	//Remove move counters and the game result at the back
	re := regexp.MustCompile(`\d+\.\s|\s1-0|\s0-1|\s1\/2-1\/2`)
	pgn.Moves = re.ReplaceAllLiteralString(pgn.Moves, "")
	SANmoves := strings.Fields(pgn.Moves)
	count := len(SANmoves)
	if count > moveLimit {
		count = moveLimit
	}
	moves := make([]string, count)
	bb := &board.Board{}
	bb.InitDefault()

	for i, san := range SANmoves {
		if i > moveLimit-1 {
			break
		}
		move := bb.SANToMove(san)
		moves[i] = move.String()
		bb.MakeMove(move)
	}

	return moves
}

func (pgn *PGN) RemoveAnnotations() {
	// Removes: move number continuation after variation `3...`, variation `(*)`, comments `{*}`, special characters `[+#?!]`
	re := regexp.MustCompile(`\d+\.\.\.|\([^()]*\)|\{[^{}]*\}|[!?+#*]`)
	whiteSpace := regexp.MustCompile(`\s+`)
	text := re.ReplaceAllLiteralString(pgn.Moves, "")
	text = whiteSpace.ReplaceAllLiteralString(text, " ")
	pgn.Moves = text

}
