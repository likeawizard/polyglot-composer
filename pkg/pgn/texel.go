package pgn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/likeawizard/polyglot-composer/pkg/board"
)

func (pgn *PGN) GetFENs() []string {
	result := "0.5"
	fens := make([]string, 0)
	switch pgn.Result {
	case "1-0":
		result = "1"
	case "0-1":
		result = "0"
	}
	annotations := pgn.GetAnnotations()

	// Based on cutechess annotations - filter out book moves and mates
	validMoveAnnotation := regexp.MustCompile(`{[+-]\d+\.\d+.*?}`)

	movesSAN := pgn.RemoveAnnotations()
	//Remove move counters and score at the end
	re := regexp.MustCompile(`\d+\.\s|\s*1-0|\s*0-1|\s*1\/2-1\/2`)
	movesSAN = re.ReplaceAllLiteralString(movesSAN, "")
	SANs := strings.Fields(movesSAN)

	b := &board.Board{}
	b.Init()

	for i, san := range SANs {
		move, err := b.SANToMove(san)
		if err != nil {
			// log.Printf("move: %s pgn: %+v\n", san, *pgn)
			break
		}

		if validMoveAnnotation.Match([]byte(annotations[i])) {
			fen := fmt.Sprintf("%s %s\n", result, b.ExportFEN())
			fens = append(fens, fen)
		}
		b.MakeMove(move)
	}
	return fens
}
