package pgn

import (
	"regexp"
)

var tagMatch = regexp.MustCompile(`\[(?P<tag>\w+)\s"(?P<value>.*[^"])"\]`)

func isTag(line string) bool {
	return tagMatch.MatchString(line)
}

func parseTag(line string) (tag Tag, value string) {
	m := tagMatch.FindStringSubmatch(line)
	tag = Tag(m[tagMatch.SubexpIndex("tag")])
	value = m[tagMatch.SubexpIndex("value")]

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
