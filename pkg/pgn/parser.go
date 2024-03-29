package pgn

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

func NewPGNParser(source Source, filtered bool) (*Parser, error) {
	pp := &Parser{
		source:   source,
		clock:    time.Now(),
		filtered: filtered,
	}

	err := pp.source.Open()
	if err != nil {
		return nil, err
	}

	return pp, nil
}

// Scan the PGN file for the next game meeting the criteria defined by filters. The game can be accessed by calling the PGN method.
func (pp *Parser) Scan(ctx context.Context) bool {
	pp.pgn = nil
	pp.tempPGN = &PGN{}
	if pp.nextLine != "" {
		pp.tempPGN.AddTag(parseTag(pp.nextLine))
		pp.nextLine = ""
	}
	for pp.source.Scan() {
		select {
		case <-ctx.Done():
			return false
		default:
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
				if pp.filtered && !pp.skipping {
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
	}

	if pp.pgn == nil && pp.tempPGN.Event != "" {
		pp.pgn = pp.tempPGN
		pp.tempPGN = nil
		pp.gameCount++
		return true
	}

	return false
}

func (pp *Parser) PGN() *PGN {
	return pp.pgn
}

func (pp *Parser) Progress(_ bool) {
	output := fmt.Sprintf("games: %d size: %v done: %.2f%%", pp.gameCount, pp.source.Size(), 100*math.Min(1, float64(pp.source.BytesRead())/float64(pp.source.Size())))
	fmt.Printf("%s\r", output)
}

func (pp *Parser) Close() error {
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

func (pgn *PGN) GetAnnotations() []string {
	re := regexp.MustCompile(`\{[^{}]*\}`)
	return re.FindAllString(pgn.Moves, -1)
}
