package pgn

import (
	"bufio"

	"github.com/dsnet/compress/bzip2"
	"github.com/inhies/go-bytesize"
)

type Bzip2PGN struct {
	reader *bzip2.Reader
	pgn    *bufio.Scanner
	close  closeFn
	path   string
	size   bytesize.ByteSize
}

func NewBzip2PGN(path string) *Bzip2PGN {
	return &Bzip2PGN{
		path: path,
	}
}

func (s *Bzip2PGN) Open() error {
	var err error
	reader, size, close, err := openSource(s.path)
	if err != nil {
		return err
	}

	s.reader, err = bzip2.NewReader(reader, nil)
	if err != nil {
		_ = close()
		return err
	}

	s.size = size
	s.close = close
	s.pgn = bufio.NewScanner(bufio.NewReader(s.reader))

	return nil
}

func (s *Bzip2PGN) Close() error {
	return s.close()
}

func (s *Bzip2PGN) Scan() bool {
	return s.pgn.Scan()
}

func (s *Bzip2PGN) Text() string {
	return s.pgn.Text()
}

func (s *Bzip2PGN) Size() bytesize.ByteSize {
	if s.reader.InputOffset > 0 {
		return s.size * bytesize.ByteSize(s.reader.OutputOffset/s.reader.InputOffset)
	}

	return s.size
}

func (s *Bzip2PGN) BytesRead() bytesize.ByteSize {
	return bytesize.ByteSize(s.reader.OutputOffset)
}
