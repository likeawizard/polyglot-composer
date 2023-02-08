package pgn

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

func NewPGNParser(source PGNSource) (*PGNParser, error) {
	pp := &PGNParser{
		source: source,
		clock:  time.Now(),
	}

	err := pp.source.Open()
	if err != nil {
		return nil, err
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
	for pp.source.Scan() {
		if time.Since(pp.clock) > time.Second {
			pp.Progress(false)
			pp.clock = time.Now()
		}
		line := pp.source.Text()
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
		pp.tempPGN = nil
		pp.gameCount++
		return true
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
	output := fmt.Sprintf("games: %d size: %v done: %.2f%%", pp.gameCount, pp.source.Size(), 100*math.Min(1, float64(pp.source.BytesRead())/float64(pp.source.Size())))
	fmt.Printf("%s\r", output)
}

func (pp *PGNParser) Close() error {
	return pp.source.Close()
}

func (pgn *PGN) RemoveAnnotations() string {
	// Removes: move number continuation after variation `3...`, variation `(*)`, comments `{*}`, special characters `[+#?!]`
	re := regexp.MustCompile(`\d+\.\.\.|\([^()]*\)|\{[^{}]*\}|[!?+#*]`)
	whiteSpace := regexp.MustCompile(`\s+`)
	text := re.ReplaceAllLiteralString(pgn.Moves, "")
	text = whiteSpace.ReplaceAllLiteralString(text, " ")
	return text
}
