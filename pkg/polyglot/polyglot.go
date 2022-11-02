package polyglot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"github.com/likeawizard/polyglot-composer/pkg/board"
	"github.com/likeawizard/polyglot-composer/pkg/pgn"
)

const (
	black_pawn = iota
	white_pawn
	black_knight
	white_knight
	black_bishop
	white_bishop
	black_rook
	white_rook
	black_queen
	white_queen
	black_king
	white_king

	entrySize = 16
	sideHash  = 780
)

type PolyglotBook map[uint64][]polyEntry

type polyEntry struct {
	move   string
	weight uint16
}

func decodeBookEntry(bytes []byte) (uint64, polyEntry) {
	key := binary.BigEndian.Uint64(bytes[:8])
	move := binary.BigEndian.Uint16(bytes[8:10])
	weight := binary.BigEndian.Uint16(bytes[10:12])

	return key, polyEntry{move: polyMoveToUCI(move), weight: weight}
}

func encodeBookEntry(key uint64, entry polyEntry) []byte {
	bytes := make([]byte, 0)
	bytes = binary.BigEndian.AppendUint64(bytes, key)
	bytes = binary.BigEndian.AppendUint16(bytes, UCIToPolyMove(entry.move))
	bytes = binary.BigEndian.AppendUint16(bytes, entry.weight)
	bytes = binary.BigEndian.AppendUint32(bytes, 0)

	return bytes
}

func polyMoveToUCI(move uint16) string {
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	promoPiece := []string{"", "n", "b", "r", "q"}
	promo := move >> 12 & 7
	fromRow := move >> 9 & 7
	fromFile := move >> 6 & 7
	toRow := move >> 3 & 7
	toFile := move & 7
	return fmt.Sprintf("%s%d%s%d%s", files[fromFile], fromRow+1, files[toFile], toRow+1, promoPiece[promo])
}

// Convert move to UCI move in general, castling moves are converted to polyglot castling move format ie e1g1 -> e1h1
func MoveToPolyMove(move board.Move) string {
	switch move {
	case board.WCastleKing:
		return "e1h1"
	case board.WCastleQueen:
		return "e1a1"
	case board.BCastleKing:
		return "e8h8"
	case board.BCastleQueen:
		return "e8a8"
	default:
		return move.String()
	}
}

func BuildFromPGNs(pgns pgn.PGNs) *PolyglotBook {
	pb := make(PolyglotBook, 0)
	for _, pgn := range pgns {
		pb.AddFromPGN(&pgn)
	}
	return &pb
}

func (pb *PolyglotBook) AddFromPGN(pgn *pgn.PGN) {
	moves := pgn.MovesToUCI()
	b := &board.Board{}
	b.Init()

	for _, move := range moves {
		ms := b.MoveGen()
		var eMove board.Move
		for _, m := range ms {
			if m.String() == move {
				eMove = m
				break
			}
		}
		switch {
		case b.Side == board.WHITE && pgn.Result == "1-0" || b.Side == board.BLACK && pgn.Result == "0-1":
			pb.AddMove(PolyZobrist(b), MoveToPolyMove(eMove), 2)
		case pgn.Result != "0-1" && pgn.Result != "1-0":
			pb.AddMove(PolyZobrist(b), MoveToPolyMove(eMove), 1)
		}
		b.MakeMove(eMove)
	}
}

func (pb *PolyglotBook) AddMove(key uint64, move string, weight uint16) {
	moves, ok := (*pb)[key]
	if ok {
		for i := 0; i < len(moves); i++ {
			if moves[i].move == move {
				moves[i].weight++
				(*pb)[key] = moves
				return
			}
		}
		moves = append(moves, polyEntry{move: move, weight: weight})
		(*pb)[key] = moves
		return
	} else {
		(*pb)[key] = []polyEntry{{move: move, weight: weight}}
	}
}

func UCIToPolyMove(move string) uint16 {
	var polyMove uint16
	files := map[byte]uint16{
		'a': 0,
		'b': 1,
		'c': 2,
		'd': 3,
		'e': 4,
		'f': 5,
		'g': 6,
		'h': 7,
	}
	ranks := map[byte]uint16{
		'1': 0,
		'2': 1,
		'3': 2,
		'4': 3,
		'5': 4,
		'6': 5,
		'7': 6,
		'8': 7,
	}
	promoPiece := map[byte]uint16{
		0:   0,
		'n': 1,
		'b': 2,
		'r': 3,
		'q': 4,
	}

	fromFile := files[move[0]]
	fromRank := ranks[move[1]]
	toFile := files[move[2]]
	toRank := ranks[move[3]]
	promo := uint16(0)
	if len(move) == 5 {
		promo = promoPiece[move[4]]
	}

	polyMove |= promo << 12
	polyMove |= fromRank << 9
	polyMove |= fromFile << 6
	polyMove |= toRank << 3
	polyMove |= toFile

	return polyMove
}

func LoadBook(path string) PolyglotBook {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("failed to open book: ", path)
	}
	polyBook := make(PolyglotBook)
	buffer := make([]byte, entrySize)
	reader := bufio.NewReader(file)

	for numBytes, err := reader.Read(buffer); err == nil && numBytes == 16; numBytes, err = reader.Read(buffer) {
		if len(buffer) != 16 {
			continue
		}
		key, entry := decodeBookEntry(buffer)
		polyBook[key] = append(polyBook[key], entry)
	}

	if err != nil {
		fmt.Println(err)
	}

	return polyBook
}

func (pb *PolyglotBook) SaveBook(path string) {
	type orderedEntry struct {
		key   uint64
		entry []polyEntry
	}
	orderedBook := make([]orderedEntry, len(*(pb)))
	n := 0

	for key, moves := range *(pb) {
		orderedBook[n] = orderedEntry{key: key, entry: moves}
		n++
	}

	sort.Slice(orderedBook, func(i, j int) bool {
		return orderedBook[i].key < orderedBook[j].key
	})

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("failed to open book: ", path)
	} else {
		defer file.Close()
	}

	writer := bufio.NewWriter(file)

	for _, moves := range orderedBook {
		for _, move := range moves.entry {
			// fmt.Println("writing")
			entry := encodeBookEntry(moves.key, move)
			// fmt.Println(len(entry))
			_, err := writer.Write(entry)
			if err != nil {
				fmt.Println("Error writing:", err)
			}
			err = writer.Flush()
			if err != nil {
				fmt.Println("Error flushing:", err)
			}
		}
	}
}

func getPieceIdx(piece, row, file int) int {
	return 64*piece + 8*row + file
}

func squareToRowAndFile(sq int) (int, int) {
	return 7 - (sq / 8), sq % 8
}

func convertPiece(piece, color int) int {
	color ^= 1
	switch piece {
	case board.PAWNS:
		piece = black_pawn
	case board.BISHOPS:
		piece = black_bishop
	case board.KNIGHTS:
		piece = black_knight
	case board.ROOKS:
		piece = black_rook
	case board.QUEENS:
		piece = black_queen
	case board.KINGS:
		piece = black_king
	}
	return piece + color
}

func PolyZobrist(b *board.Board) uint64 {
	var hash uint64
	for color := board.WHITE; color <= board.BLACK; color++ {
		for piece := board.PAWNS; piece <= board.KINGS; piece++ {
			polyPiece := convertPiece(piece, color)
			pieces := b.Pieces[color][piece]
			for pieces > 0 {
				sq := pieces.PopLS1B()
				row, file := squareToRowAndFile(sq)
				hash ^= zobristHashes[getPieceIdx(polyPiece, row, file)]
			}
		}
	}

	cRs := [4]board.CastlingRights{board.WOO, board.WOOO, board.BOO, board.BOOO}
	polyCastling := [4]int{768, 769, 770, 771}

	for idx, cr := range cRs {
		if b.CastlingRights&cr != 0 {
			hash ^= zobristHashes[polyCastling[idx]]
		}
	}

	if b.EnPassantTarget > 0 && b.Pieces[b.Side][board.PAWNS]&board.PawnAttacks[b.Side^1][b.EnPassantTarget] != 0 {
		hash ^= zobristHashes[772+b.EnPassantTarget%8]
	}

	if b.Side == board.WHITE {
		hash ^= zobristHashes[sideHash]
	}

	return hash
}
