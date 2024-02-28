package pgn

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/likeawizard/tofiks/pkg/board"
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

	re := regexp.MustCompile(`\d+\.\s|\s*1-0|\s*0-1|\s*1\/2-1\/2`)
	movesSAN = re.ReplaceAllLiteralString(movesSAN, "")
	SANs := strings.Fields(movesSAN)

	b := board.NewBoard("startpos")

	for i, san := range SANs {
		move, err := SANToMove(b, san)
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

func SANToMove(b *board.Board, san string) (board.Move, error) {
	switch {
	// Castling moves
	case san == "O-O-O":
		if b.Side == board.WHITE {
			return board.WCastleQueen, nil
		} else {
			return board.BCastleQueen, nil
		}
	case san == "O-O":
		if b.Side == board.WHITE {
			return board.WCastleKing, nil
		} else {
			return board.BCastleKing, nil
		}
	default:
		sanRe := regexp.MustCompile(`^(?P<piece>[NBRQK])?(?P<disamb>[a-h]?[1-8]?)?(?P<capture>x?)(?P<target>[a-h][1-8])(?:=(?P<promo>[NBRQ]))?$`)
		m := sanRe.FindStringSubmatch(san)
		piece := m[sanRe.SubexpIndex("piece")]
		disamb := m[sanRe.SubexpIndex("disamb")]
		target := m[sanRe.SubexpIndex("target")]
		promo := m[sanRe.SubexpIndex("promo")]
		return getMoveWithFromTo(b, piece, target, disamb, promo)
	}
}

func getMoveWithFromTo(b *board.Board, pieceStr, to, dis, promo string) (board.Move, error) {
	moves := b.MoveGenLegal()
	piece := board.PAWNS
	if pieceStr != "" {
		switch pieceStr {
		case "B":
			piece = board.BISHOPS
		case "N":
			piece = board.KNIGHTS
		case "R":
			piece = board.ROOKS
		case "Q":
			piece = board.QUEENS
		case "K":
			piece = board.KINGS
		}
	}
	if b.Side == board.WHITE {
		piece++
	} else {
		piece += 7
	}

	for _, move := range moves {
		movePromo := ""
		if len(move.String()) == 5 {
			movePromo = strings.ToUpper(move.String()[4:])
		}
		if dis != "" {
			if len(dis) == 2 {
				if int(move.Piece()) == piece && move.To().String() == to && move.From().String() == dis && movePromo == promo {
					return move, nil
				}
			} else {
				rank, err := strconv.Atoi(dis)
				if err == nil {
					rank = 7 - (rank - 1)
					if int(move.Piece()) == piece && move.To().String() == to && int(move.From())/8 == rank && movePromo == promo {
						return move, nil
					}
				} else {
					files := map[string]int{
						"a": 0,
						"b": 1,
						"c": 2,
						"d": 3,
						"e": 4,
						"f": 5,
						"g": 6,
						"h": 7,
					}
					file := files[dis]
					if int(move.Piece()) == piece && move.To().String() == to && int(move.From())%8 == file && movePromo == promo {
						return move, nil
					}
				}
			}
		} else if int(move.Piece()) == piece && move.To().String() == to {
			return move, nil
		}
	}
	return 0, fmt.Errorf("unable to convert SAN to move with: from '%s', to '%s', dis '%s', promo '%s'", pieceStr, to, dis, promo)
}
