package pgn

import (
	"bufio"

	"github.com/inhies/go-bytesize"
)

type PlainPGN struct {
	pgn    *bufio.Scanner
	reader *ByteCountingReader
	close  closeFn
	path   string
	size   bytesize.ByteSize
}

func NewPlainPGN(path string) *PlainPGN {
	return &PlainPGN{
		path: path,
	}
}

func (s *PlainPGN) Open() error {
	var err error
	reader, size, close, err := openSource(s.path)
	if err != nil {
		return err
	}

	s.close = close
	s.reader = &ByteCountingReader{reader: reader}
	s.size = size
	s.pgn = bufio.NewScanner(bufio.NewReader(s.reader))

	return nil
}

func (s *PlainPGN) Close() error {
	return s.close()
}

func (s *PlainPGN) Scan() bool {
	return s.pgn.Scan()
}

func (s *PlainPGN) Text() string {
	return s.pgn.Text()
}

func (s *PlainPGN) Size() bytesize.ByteSize {
	return s.size
}

func (s *PlainPGN) BytesRead() bytesize.ByteSize {
	return s.reader.bytesRead
}
