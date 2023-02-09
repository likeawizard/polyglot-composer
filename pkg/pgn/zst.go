package pgn

import (
	"bufio"
	"os"

	"github.com/inhies/go-bytesize"
	"github.com/klauspost/compress/zstd"
)

type ZstPGN struct {
	path         string
	file         *os.File
	pgn          *bufio.Scanner
	inputReader  *ByteCountingReader
	outputReader *ByteCountingReader
	size         bytesize.ByteSize
	close        closeFn
}

func NewZstPGN(path string) *ZstPGN {
	return &ZstPGN{
		path: path,
	}
}

func (s *ZstPGN) Open() error {
	var err error
	reader, size, close, err := openSource(s.path)
	if err != nil {
		return err
	}
	// Wrap the file and the zstd Reader in ByteCountingReader to estimate the data size by output/input ratio
	s.inputReader = &ByteCountingReader{reader: reader}
	zstReader, err := zstd.NewReader(s.inputReader)
	if err != nil {
		return err
	}

	s.outputReader = &ByteCountingReader{reader: zstReader}

	s.close = close
	s.size = size
	s.pgn = bufio.NewScanner(bufio.NewReader(s.outputReader))

	return nil
}

func (s *ZstPGN) Close() error {
	return s.file.Close()
}

func (s *ZstPGN) Scan() bool {
	return s.pgn.Scan()
}

func (s *ZstPGN) Text() string {
	return s.pgn.Text()
}

func (s *ZstPGN) Size() bytesize.ByteSize {
	if s.inputReader.bytesRead > 0 {
		return s.size * (s.outputReader.bytesRead / s.inputReader.bytesRead)
	}

	return s.size
}

func (s *ZstPGN) BytesRead() bytesize.ByteSize {
	return s.outputReader.bytesRead
}
