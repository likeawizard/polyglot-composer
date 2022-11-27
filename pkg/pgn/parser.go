package pgn

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"

	"github.com/inhies/go-bytesize"
)

func NewPGNParser(path string) (*PGNParser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening PGN: %s", err)
	}

	size := float64(1)
	stat, err := file.Stat()
	if err == nil {
		size = float64(stat.Size())
	}

	pp := &PGNParser{
		clock:      time.Now(),
		file:       file,
		totalBytes: bytesize.New(size)}

	err = pp.setReader()
	if err != nil {
		return nil, err
	}

	return pp, nil
}

func (pp *PGNParser) setReader() error {
	stat, err := pp.file.Stat()
	if err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(stat.Name(), "zst"):
		pp.isArchived = true
		zstReader, err := zstd.NewReader(pp.file)
		if err != nil {
			return fmt.Errorf("error opening zst: %s", err)
		}
		pp.source = zstReader
		pp.scanner = bufio.NewScanner(bufio.NewReader(zstReader))

	case strings.HasSuffix(stat.Name(), "bz2"):
		pp.isArchived = true
		bzReader, err := bzip2.NewReader(pp.file, nil)
		if err != nil {
			return fmt.Errorf("error opening bzip2: %s", err)
		}
		pp.source = bzReader
		pp.scanner = bufio.NewScanner(bufio.NewReader(bzReader))

	default:
		pp.source = pp.file
		pp.scanner = bufio.NewScanner(bufio.NewReader(pp.file))
	}

	return nil
}

// Scan the PGN file for the next game meeting the criteria defined by filters. The game can be accessed by calling the PGN method
func (pp *PGNParser) Scan() bool {
	pp.pgn = nil
	pp.tempPGN = &PGN{}
	if pp.nextLine != "" {
		pp.tempPGN.AddTag(parseTag(pp.nextLine))
		pp.nextLine = ""
	}
	for pp.scanner.Scan() {
		pp.lastBytes += bytesize.New(float64(len(pp.scanner.Bytes())))
		if time.Since(pp.clock) > time.Second {
			pp.Progress(false)
			pp.readBytes += pp.lastBytes
			pp.lastBytes = 0
			pp.clock = time.Now()
		}
		line := pp.scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if isTag(line) {
			pp.tag, pp.value = parseTag(line)
			if !pp.skipping {
				pp.skipping = !PreFilter(pp.tag, pp.value)
			}
			if pp.tag == TAG_EVENT && pp.tempPGN.Event != "" {
				if pp.skipping {
					pp.tempPGN = &PGN{Event: pp.value}
					pp.skipping = false
				} else {
					pp.pgn = pp.tempPGN
					pp.nextLine = line
					pp.gameCount++
					return true
				}
			}

			pp.tempPGN.AddTag(pp.tag, pp.value)

		} else {
			pp.tempPGN.Moves += line
		}
	}

	if pp.pgn == nil && pp.tempPGN.Event != "" {
		pp.pgn = pp.tempPGN
		pp.tempPGN = nil
		pp.gameCount++
		return true
	}

	return false
}

func (pp *PGNParser) PGN() *PGN {
	return pp.pgn

}

func (pp *PGNParser) Next() *PGN {
	pp.Scan()
	return pp.pgn
}

func (pp *PGNParser) Progress(done bool) {
	ratio := pp.ratio()
	progress := math.Min(float64(pp.readBytes)/float64(pp.totalBytes)/ratio, 1)
	if done {
		progress = 1
	}
	barN := int(50 * progress)
	bar := "[" + strings.Repeat("#", barN) + strings.Repeat(".", 50-barN) + "]"
	output := fmt.Sprintf("%s %.2f%%, games: %d rate: %v/s read: %v, total: %v", bar, 100*progress, pp.gameCount, pp.lastBytes, pp.readBytes, pp.totalBytes*bytesize.New(ratio))
	if pp.isArchived {
		output += " (estimate)                       "
	}

	fmt.Printf("%s\r", output)
}

func (pp *PGNParser) ratio() float64 {
	switch reader := pp.source.(type) {
	case *zstd.Decoder:
		//to do
		return 8.0
	case *bzip2.Reader:
		if reader.InputOffset > 0 {
			return float64(reader.OutputOffset) / float64(reader.InputOffset)
		} else {
			return 1.0
		}
	default:
		return 1.0
	}
}

func (pp *PGNParser) Close() error {
	return pp.file.Close()
}

func (pgn *PGN) RemoveAnnotations() string {
	// Removes: move number continuation after variation `3...`, variation `(*)`, comments `{*}`, special characters `[+#?!]`
	re := regexp.MustCompile(`\d+\.\.\.|\([^()]*\)|\{[^{}]*\}|[!?+#*]`)
	whiteSpace := regexp.MustCompile(`\s+`)
	text := re.ReplaceAllLiteralString(pgn.Moves, "")
	text = whiteSpace.ReplaceAllLiteralString(text, " ")
	return text

}
