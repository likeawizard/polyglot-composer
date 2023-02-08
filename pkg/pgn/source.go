package pgn

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/inhies/go-bytesize"
	"github.com/klauspost/compress/zstd"
)

type PGNSource interface {
	Open() error
	Close() error
	Scan() bool
	Text() string
	// The size (or estimated size in case of archives) of the data
	Size() bytesize.ByteSize
	// Fraction of data that has been read. Bounded by 1 in case of bad estimates
	BytesRead() bytesize.ByteSize
}

// Reader that keeps track on the bytes read.
// Useful for reading compressed files and estimateing their compression by wrapping input and output readers (i.e. input: file, output: decompressed data read)
type ByteCountingReader struct {
	bytesRead bytesize.ByteSize
	reader    io.Reader
}

func (bcr *ByteCountingReader) Read(p []byte) (n int, err error) {
	c, err := bcr.reader.Read(p)
	bcr.bytesRead += bytesize.ByteSize(uint64(c))
	return c, err
}

func sourceFromPath(fileName string) (PGNSource, error) {
	switch {
	case strings.HasSuffix(fileName, ".zst"):
		return NewZstPGN(fileName), nil
	case strings.HasSuffix(fileName, ".bz2"):
		return NewBzip2PGN(fileName), nil
	case strings.HasSuffix(fileName, ".pgn"):
		return NewPlainPGN(fileName), nil
	default:
		return nil, fmt.Errorf("unsupported file format")
	}
}

func ParsePath(pgnPath string) ([]PGNSource, error) {
	var files []PGNSource
	paths := strings.Split(pgnPath, ",")

	for _, path := range paths {
		path = strings.TrimSpace(path)
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("error opening PGN: %s", err)
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("error getting file stats: %s", err)
		}

		if fileInfo.IsDir() {
			subfiles, err := os.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("error reading directory: %s", err)
			}

			for _, subfile := range subfiles {
				if source, err := sourceFromPath(filepath.Join(path, subfile.Name())); err == nil {
					files = append(files, source)
				}
			}
		} else {
			if source, err := sourceFromPath(path); err == nil {
				files = append(files, source)
			}
		}
	}
	return files, nil
}

type PlainPGN struct {
	path   string
	file   *os.File
	pgn    *bufio.Scanner
	reader *ByteCountingReader
	size   bytesize.ByteSize
}

func NewPlainPGN(path string) *PlainPGN {
	return &PlainPGN{
		path: path,
	}
}

func (s *PlainPGN) Open() error {
	var err error
	s.file, err = os.Open(s.path)
	if err != nil {
		return err
	}

	s.reader = &ByteCountingReader{reader: s.file}

	stat, err := s.file.Stat()
	if err != nil {
		return err
	}

	s.size = bytesize.New(float64(stat.Size()))
	s.pgn = bufio.NewScanner(bufio.NewReader(s.reader))

	return nil
}

func (s *PlainPGN) Close() error {
	return s.file.Close()
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

type Bzip2PGN struct {
	path   string
	file   *os.File
	reader *bzip2.Reader
	pgn    *bufio.Scanner
	size   bytesize.ByteSize
}

func NewBzip2PGN(path string) *Bzip2PGN {
	return &Bzip2PGN{
		path: path,
	}
}

func (s *Bzip2PGN) Open() error {
	var err error
	s.file, err = os.Open(s.path)
	if err != nil {
		return err
	}

	s.reader, err = bzip2.NewReader(s.file, nil)
	if err != nil {
		return err
	}

	stat, err := s.file.Stat()
	if err != nil {
		return err
	}

	s.size = bytesize.New(float64(stat.Size()))
	s.pgn = bufio.NewScanner(bufio.NewReader(s.reader))

	return nil
}

func (s *Bzip2PGN) Close() error {
	return s.file.Close()
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

type ZstPGN struct {
	path         string
	file         *os.File
	pgn          *bufio.Scanner
	inputReader  *ByteCountingReader
	outputReader *ByteCountingReader
	size         bytesize.ByteSize
}

func NewZstPGN(path string) *ZstPGN {
	return &ZstPGN{
		path: path,
	}
}

func (s *ZstPGN) Open() error {
	var err error
	s.file, err = os.Open(s.path)
	if err != nil {
		return err
	}

	// Wrap the file and the zstd Reader in ByteCountingReader to estimate the data size by output/input ratio
	s.inputReader = &ByteCountingReader{reader: s.file}
	zstReader, err := zstd.NewReader(s.inputReader)
	if err != nil {
		return err
	}

	s.outputReader = &ByteCountingReader{reader: zstReader}

	stat, err := s.file.Stat()
	if err != nil {
		return err
	}

	s.size = bytesize.New(float64(stat.Size()))
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
		return s.size * s.outputReader.bytesRead / s.inputReader.bytesRead
	}

	return s.size
}

func (s *ZstPGN) BytesRead() bytesize.ByteSize {
	return s.outputReader.bytesRead
}
