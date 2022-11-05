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
	case TAG_RESULT:
		pgn.Result = value
	}

	return pgn
}
