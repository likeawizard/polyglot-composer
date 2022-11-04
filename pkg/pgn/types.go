package pgn

import (
	"bufio"
	"os"
	"time"

	"github.com/dsnet/compress/bzip2"

	"github.com/inhies/go-bytesize"
)

const (
	TERM_NORMAL = "Normal"
	TERM_TIME   = "Time forfeit"
)

// Upper bounds for adjusted time per game (seconds), bullet assumes 40move game, blitz, rapid 60 move game
const (
	TC_BULLET    = 0
	TC_BLITZ     = 3 * 60
	TC_RAPID     = 10 * 60
	TC_CLASSICAL = 60 * 60
)

type Tag string

const (
	TAG_EVENT       Tag = "Event"
	TAG_SITE        Tag = "Site"
	TAG_DATE        Tag = "Date"
	TAG_ROUND       Tag = "Round"
	TAG_WHITE       Tag = "White"
	TAG_BLACK       Tag = "Black"
	TAG_RESULT      Tag = "Result"
	TAG_TERMINATION Tag = "Termination"
	TAG_TIMECONTROL Tag = "TimeControl"
	TAG_ECO         Tag = "ECO"
	TAG_WHITE_ELO   Tag = "WhiteElo"
	TAG_BLACK_ELO   Tag = "BlackElo"
)

type PGNs []PGN

type PGN struct {
	Event       string
	Site        string
	Date        string
	White       string
	Black       string
	Result      string
	UTCDate     string
	UTCTime     string
	WhiteElo    string
	BlackElo    string
	Variant     string
	TimeControl string
	ECO         string
	Termination string
	Moves       string
}

type PGNParser struct {
	clock      time.Time
	file       *os.File
	bzipReader *bzip2.Reader
	isArchived bool
	skipping   bool
	gameCount  int
	totalBytes bytesize.ByteSize
	readBytes  bytesize.ByteSize
	lastBytes  bytesize.ByteSize
	scanner    *bufio.Scanner
	pgn        *PGN
	nextLine   string
}
