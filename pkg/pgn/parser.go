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

// Scan the PGN file for the next game meeting the criteria defined by filters. The game can be accessed by calling the PGN method
func (pp *PGNParser) Scan() bool {
	pp.pgn = nil
	pp.tempPGN = &PGN{}
	if pp.nextLine != "" {
		pp.tempPGN.AddTag(parseTag(pp.nextLine))
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
			pp.tag, pp.value = parseTag(line)
			if !pp.skipping {
				pp.skipping = !PreFilter(pp.tag, pp.value)
			}
			if pp.tag == TAG_EVENT && pp.tempPGN.Event != "" {
				if pp.skipping {
					pp.tempPGN = &PGN{Event: pp.value}
					pp.skipping = false
				} else {
					pp.pgn = pp.tempPGN
					pp.nextLine = line
					pp.gameCount++
					return true
				}
			}

			pp.tempPGN.AddTag(pp.tag, pp.value)

		} else {
			pp.tempPGN.Moves += line
		}
	}

	if pp.pgn == nil && pp.tempPGN.Event != "" {
		pp.pgn = pp.tempPGN
		pp.gameCount++
	}

	return false
}

func (pp *PGNParser) PGN() *PGN {
	return pp.pgn

}

func (pp *PGNParser) Next() *PGN {
	pp.Scan()
	return pp.pgn
}

func (pp *PGNParser) Progress(done bool) {
	ratio := 1.0
	if pp.bzipReader != nil && pp.bzipReader.InputOffset > 0 && pp.bzipReader.OutputOffset > 0 {
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
		output += " (estimate)                       "
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
	moves := make([]string, 0)
	bb := &board.Board{}
	bb.InitDefault()

	for i, san := range SANmoves {
		if i > moveLimit-1 {
			break
		}
		move, err := bb.SANToMove(san)
		if err != nil {
			break
		}
		moves = append(moves, move.String())
		bb.MakeMove(move)
	}

	return moves
}

func (pgn *PGN) RemoveAnnotations() string {
	// Removes: move number continuation after variation `3...`, variation `(*)`, comments `{*}`, special characters `[+#?!]`
	re := regexp.MustCompile(`\d+\.\.\.|\([^()]*\)|\{[^{}]*\}|[!?+#*]`)
	whiteSpace := regexp.MustCompile(`\s+`)
	text := re.ReplaceAllLiteralString(pgn.Moves, "")
	text = whiteSpace.ReplaceAllLiteralString(text, " ")
	return text

}
