package pgn

import (
	"time"
)

const (
	TERM_NORMAL = "Normal"
	TERM_TIME   = "Time forfeit"
)

// Upper bounds for adjusted time per game (seconds), bullet assumes 40move game, blitz, rapid 60 move game.
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
	Event string
	// White       string
	// Black       string
	Result string
	// WhiteElo    string
	// BlackElo    string
	// ECO         string
	Moves string
}

type Parser struct {
	clock     time.Time
	source    Source
	pgn       *PGN
	tempPGN   *PGN
	tag       Tag
	value     string
	nextLine  string
	gameCount int
	skipping  bool
	filtered  bool
}
