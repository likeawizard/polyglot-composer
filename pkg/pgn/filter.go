package pgn

import (
	"strconv"
	"strings"
)

// Filter conditions that can be determined on individual Tags
func PreFilter(tag Tag, value string) bool {
	switch tag {
	case TAG_RESULT:
		return value == "1-0" || value == "0-1" || value == "1/2-1/2"
	case TAG_TERMINATION:
		return value == TERM_NORMAL
	case TAG_TIMECONTROL:
		return TC_BLITZ <= adjustedTime(value)
	case TAG_WHITE_ELO, TAG_BLACK_ELO:
		elo, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		return elo > 2500
	default:
		return true
	}
}

// Filter conditions that dependant on a combination of tags. ie Elo difference
// TODO: function stub
func (pgn *PGN) PostFilter() bool {
	return true
}

func adjustedTime(value string) int {
	parts := strings.Split(value, "+")
	partsInt := make([]int, len(parts))
	var err error
	for i, part := range parts {
		partsInt[i], err = strconv.Atoi(part)
		if err != nil {
			return 0
		}
	}
	switch len(partsInt) {
	case 1:
		return partsInt[0]
	case 2:
		return partsInt[0] + 60*partsInt[1]
	default:
		return 0
	}
}
