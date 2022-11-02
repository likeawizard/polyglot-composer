package pgn

import (
	"bufio"
	"compress/bzip2"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/likeawizard/polyglot-composer/pkg/board"
)

type Termination string

const (
	T_NORMAL Termination = "Normal"
	T_TIME   Termination = "Time forfeit"
)

type Tag string

const (
	TAG_EVENT  Tag = "Event"
	TAG_SITE   Tag = "Site"
	TAG_DATE   Tag = "Date"
	TAG_ROUND  Tag = "Round"
	TAG_WHITE  Tag = "White"
	TAG_BLACK  Tag = "Black"
	TAG_RESULT Tag = "Result"
)

type PGNs []PGN

type PGN struct {
	Event string

	Site            string
	Date            string
	White           string
	Black           string
	Result          string
	UTCDate         string
	UTCTime         string
	WhiteElo        string
	BlackElo        string
	WhiteRatingDiff string
	BlackRatingDiff string
	Variant         string
	TimeControl     string
	ECO             string
	Termination     string
	Moves           string
}

type PGNParser struct {
	file     *os.File
	scanner  *bufio.Scanner
	pgn      *PGN
	nextLine string
}

func NewPGNParser(path string) (*PGNParser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening PGN:", err)
	}

	var scanner *bufio.Scanner
	if strings.HasSuffix(path, "bz2") {
		bzReader := bzip2.NewReader(file)
		scanner = bufio.NewScanner(bufio.NewReader(bzReader))
	} else {
		scanner = bufio.NewScanner(bufio.NewReader(file))
	}

	return &PGNParser{scanner: scanner, file: file}, nil
}

func (pp *PGNParser) Scan() {
	pp.pgn = nil
	pgn := PGN{}
	if pp.nextLine != "" {
		pgn.AddTag(parseTag(pp.nextLine))
		pp.nextLine = ""
	}
	for pp.scanner.Scan() {
		line := pp.scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if isTag(line) {
			tag, value := parseTag(line)
			if tag == TAG_EVENT && pgn.Event != "" {
				pp.pgn = &pgn
				pp.nextLine = line
				return
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

	return pp.pgn

}

func (pp *PGNParser) Close() error {
	return pp.file.Close()
}

func ParsePGN(path string) PGNs {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening PGN:", err)
	}
	_ = file

	pgns := make(PGNs, 0)
	var pgn PGN

	scanner := bufio.NewScanner(bufio.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if isTag(line) {
			tag, value := parseTag(line)
			if tag == TAG_EVENT && pgn.Event != "" {
				pgn.RemoveAnnotations()
				pgns = append(pgns, pgn)
				pgn = PGN{}
			}

			pgn.AddTag(tag, value)

		} else {
			pgn.Moves += line
		}
	}

	if pgn.Event != "" {
		pgn.RemoveAnnotations()
		pgns = append(pgns, pgn)
		pgn = PGN{}
	}

	return pgns
}

func isTag(line string) bool {
	return len(line) > 2 && line[0] == '[' && line[len(line)-1] == ']'
}

func parseTag(line string) (tag Tag, value string) {
	// cut the opening [ and closing "]"
	line = line[1 : len(line)-2]
	fields := strings.Fields(line)
	tag = Tag(fields[0])
	// cut the opening value double quote
	value = fields[1][1:]
	return
}

func (pgn *PGN) AddTag(tag Tag, value string) *PGN {
	switch tag {
	case TAG_EVENT:
		pgn.Event = value
	case TAG_SITE:
		pgn.Site = value
	case TAG_DATE:
		pgn.Date = value
	case TAG_WHITE:
		pgn.White = value
	case TAG_BLACK:
		pgn.Black = value
	case TAG_RESULT:
		pgn.Result = value
	}

	return pgn
}

func (pgn *PGN) MovesToUCI() []string {
	moves := make([]string, 0)
	bb := &board.Board{}
	bb.InitDefault()
	SANmoves := strings.Fields(pgn.Moves)
	for _, san := range SANmoves {
		_, err := strconv.Atoi(san[:1])
		// if move begins with an integer it's either move counter or reslt (1-0, 0-1, 1/2-1/2)
		if err == nil {
			continue
		}
		move := bb.SANToMove(san)
		moves = append(moves, move.String())
		bb.MakeMove(move)
	}

	return moves
}

func (pgn *PGN) RemoveAnnotations() {
	// Removes: move number continuation after variation `3...`, variation `(*)`, comments `{*}`, special characters `[+#?!]`
	re := regexp.MustCompile(`\d+\.\.\.|\([^()]*\)|\{[^{}]*\}|[!?+#]`)
	whiteSpace := regexp.MustCompile(`\s+`)
	empty := ""
	text := re.ReplaceAll([]byte(pgn.Moves), []byte(empty))
	text = whiteSpace.ReplaceAll(text, []byte(" "))
	pgn.Moves = string(text)

}
