package pgn

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/inhies/go-bytesize"
)

type Source interface {
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
// Useful for reading compressed files and estimateing their compression by wrapping input and output readers (i.e. input: file, output: decompressed data read).
type ByteCountingReader struct {
	reader    io.Reader
	bytesRead bytesize.ByteSize
}

func (bcr *ByteCountingReader) Read(p []byte) (n int, err error) {
	c, err := bcr.reader.Read(p)
	bcr.bytesRead += bytesize.ByteSize(uint64(c))
	return c, err
}

type closeFn func() error

func openSource(path string) (io.Reader, bytesize.ByteSize, closeFn, error) {
	if isUrl(path) {
		r, err := http.Get(path)
		if err != nil {
			return nil, 0, nil, err
		}

		return r.Body, bytesize.ByteSize(r.ContentLength), r.Body.Close, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, 0, nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, nil, err
	}

	return file, bytesize.ByteSize(stat.Size()), file.Close, err
}

func sourceFromPath(path string) (Source, error) {
	switch filepath.Ext(path) {
	case ".zst":
		return NewZstPGN(path), nil
	case ".bz2":
		return NewBzip2PGN(path), nil
	case ".pgn":
		return NewPlainPGN(path), nil
	default:
		return nil, fmt.Errorf("unsupported file format")
	}
}

func ParsePath(pgnPath string) ([]Source, error) {
	files := make([]Source, 0)
	paths := strings.Split(pgnPath, ",")

	for _, path := range paths {
		path = strings.TrimSpace(path)

		if isUrl(path) {
			if source, err := sourceFromPath(path); err == nil {
				files = append(files, source)
			}
		} else {
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			defer file.Close()

			fileInfo, err := file.Stat()
			if err != nil {
				continue
			}

			if fileInfo.IsDir() {
				if dirSources, err := dataSourceFromDir(path); err == nil {
					files = append(files, dirSources...)
				}
			} else {
				if source, err := sourceFromPath(path); err == nil {
					files = append(files, source)
				}
			}
		}
	}
	return files, nil
}

func isUrl(path string) bool {
	_, err := url.ParseRequestURI(path)
	return err == nil
}

func dataSourceFromDir(path string) ([]Source, error) {
	subfiles, err := os.ReadDir(path)
	sc := make([]Source, 0)
	if err != nil {
		return nil, err
	}

	for _, subfile := range subfiles {
		// Ignore subdirectories
		if subfile.IsDir() {
			continue
		}

		if source, err := sourceFromPath(filepath.Join(path, subfile.Name())); err == nil {
			sc = append(sc, source)
		}
	}
	return sc, nil
}
