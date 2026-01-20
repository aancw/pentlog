package logs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ttyrecHeader struct {
	Sec  uint32
	Usec uint32
	Len  uint32
}

func MergeTTYFiles(files []string, outputFile string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to merge")
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return files[i] < files[j]
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	var outBuf bytes.Buffer
	var globalOffsetSec uint32 = 0
	var globalOffsetUsec uint32 = 0

	for idx, file := range files {
		if idx > 0 {
			clearScreen := "\033[2J\033[H"
			banner := fmt.Sprintf("\n\n[ --- Starting Session: %s --- ]\n\n", filepath.Base(file))
			separatorData := clearScreen + banner

			writeRecord(&outBuf, globalOffsetSec, globalOffsetUsec, []byte(separatorData))

			globalOffsetSec += 2
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		var lastRecordSec uint32 = 0
		var lastRecordUsec uint32 = 0
		var firstRecord bool = true
		var baseOffsetSec uint32 = 0
		var baseOffsetUsec uint32 = 0

		reader := bytes.NewReader(data)
		for reader.Len() >= 12 {
			var header ttyrecHeader
			if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
				break
			}

			if reader.Len() < int(header.Len) {
				break
			}

			recordData := make([]byte, header.Len)
			if _, err := reader.Read(recordData); err != nil {
				break
			}

			if firstRecord {
				baseOffsetSec = header.Sec
				baseOffsetUsec = header.Usec
				firstRecord = false
			}

			relativeSec := header.Sec - baseOffsetSec
			relativeUsec := header.Usec
			if header.Usec < baseOffsetUsec && relativeSec > 0 {
				relativeSec--
				relativeUsec = 1000000 + header.Usec - baseOffsetUsec
			} else {
				relativeUsec = header.Usec - baseOffsetUsec
			}

			newSec := globalOffsetSec + relativeSec
			newUsec := globalOffsetUsec + relativeUsec
			if newUsec >= 1000000 {
				newSec++
				newUsec -= 1000000
			}

			writeRecord(&outBuf, newSec, newUsec, recordData)

			lastRecordSec = relativeSec
			lastRecordUsec = relativeUsec
		}

		globalOffsetSec += lastRecordSec + 1
		globalOffsetUsec = lastRecordUsec
		if globalOffsetUsec >= 1000000 {
			globalOffsetSec++
			globalOffsetUsec -= 1000000
		}
	}

	return os.WriteFile(outputFile, outBuf.Bytes(), 0644)
}

func writeRecord(buf *bytes.Buffer, sec, usec uint32, data []byte) {
	header := ttyrecHeader{
		Sec:  sec,
		Usec: usec,
		Len:  uint32(len(data)),
	}
	binary.Write(buf, binary.LittleEndian, header)
	buf.Write(data)
}
