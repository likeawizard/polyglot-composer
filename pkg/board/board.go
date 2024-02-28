package board

var Pieces = [6]string{"P", "B", "N", "R", "Q", "K"}

func (b *Board) Init() {
	fen := startingFEN
	_ = b.ImportFEN(fen)
}

func (b *Board) InitDefault() {
	_ = b.ImportFEN(startingFEN)
}

func (b *Board) Copy() *Board {
	copy := Board{
		Hash:            b.Hash,
		Pieces:          b.Pieces,
		Occupancy:       b.Occupancy,
		Side:            b.Side,
		CastlingRights:  b.CastlingRights,
		EnPassantTarget: b.EnPassantTarget,
		HalfMoveCounter: b.HalfMoveCounter,
		FullMoveCounter: b.FullMoveCounter,
	}

	return &copy
}

func (b *Board) updateCastlingRights(move Move) {
	if b.CastlingRights == 0 {
		return
	}
	from, to := move.FromTo()

	switch {
	case b.CastlingRights&(WOOO|WOO) != 0 && from == WCastleQueen.From():
		b.CastlingRights &^= WOOO
		b.CastlingRights &^= WOO
	case b.CastlingRights&(BOOO|BOO) != 0 && from == BCastleQueen.From():
		b.CastlingRights &^= BOOO
		b.CastlingRights &^= BOO
	case b.CastlingRights&WOOO != 0 && (from == WCastleQueenRook.From() || to == WCastleQueenRook.From()):
		b.CastlingRights &^= WOOO
	case b.CastlingRights&WOO != 0 && (from == WCastleKingRook.From() || to == WCastleKingRook.From()):
		b.CastlingRights &^= WOO
	case b.CastlingRights&BOOO != 0 && (from == BCastleQueenRook.From() || to == BCastleQueenRook.From()):
		b.CastlingRights &^= BOOO
	case b.CastlingRights&BOO != 0 && (from == BCastleKingRook.From() || to == BCastleKingRook.From()):
		b.CastlingRights &^= BOO
	}
}
