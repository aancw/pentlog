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

func InsertResumeMarker(ttyPath string) error {
	data, err := os.ReadFile(ttyPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var lastSec, lastUsec uint32
	reader := bytes.NewReader(data)

	for reader.Len() >= 12 {
		var header ttyrecHeader
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			break
		}
		if reader.Len() < int(header.Len) {
			break
		}
		lastSec = header.Sec
		lastUsec = header.Usec
		reader.Seek(int64(header.Len), 1)
	}

	f, err := os.OpenFile(ttyPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for append: %w", err)
	}
	defer f.Close()

	banner := "\r\n\r\n═══════════════════════════════════════════════════════════════\r\n"
	banner += "                    Session Resumed\r\n"
	banner += "═══════════════════════════════════════════════════════════════\r\n\r\n"

	var buf bytes.Buffer

	// Banner at lastSec + 2 (leaving a small gap)
	markerSec := lastSec + 2
	markerUsec := lastUsec

	writeRecord(&buf, markerSec, markerUsec, []byte(banner))

	_, err = f.Write(buf.Bytes())
	return err
}

func GetLastFrameTimestamp(ttyPath string) (uint32, error) {
	data, err := os.ReadFile(ttyPath)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	var lastSec uint32
	reader := bytes.NewReader(data)

	for reader.Len() >= 12 {
		var header ttyrecHeader
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			break
		}
		if reader.Len() < int(header.Len) {
			break
		}
		lastSec = header.Sec
		reader.Seek(int64(header.Len), 1)
	}

	return lastSec, nil
}

func AdjustFutureTimestamps(ttyPath string, timeDelta int64) error {
	resumeMarkerPath := ttyPath + ".resume_offset"
	offsetData := fmt.Sprintf("%d", timeDelta)
	return os.WriteFile(resumeMarkerPath, []byte(offsetData), 0644)
}

func NormalizeResumedSession(ttyPath string) error {
	data, err := os.ReadFile(ttyPath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}
	type frame struct {
		sec     uint32
		usec    uint32
		payload []byte
	}

	var frames []frame
	var resumeBannerIdx int = -1

	// Parse all frames from the tty file
	reader := bytes.NewReader(data)
	for reader.Len() >= 12 {
		var header ttyrecHeader
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			break
		}
		if reader.Len() < int(header.Len) {
			break
		}

		payload := make([]byte, header.Len)
		if _, err := reader.Read(payload); err != nil {
			break
		}

		frames = append(frames, frame{
			sec:     header.Sec,
			usec:    header.Usec,
			payload: payload,
		})

		// Track the index of the resume banner frame
		// Search for "Resumed" to match both old (ANSI) and new (plain) banners
		if bytes.Contains(payload, []byte("Resumed")) {
			resumeBannerIdx = len(frames) - 1
		}
	}

	// No resume banner found - nothing to normalize
	if resumeBannerIdx < 0 {
		return nil
	}

	// Find the timestamp of the last frame BEFORE the resume banner
	// This is the last frame from the original (crashed) session
	var lastBeforeResumeSec uint32
	if resumeBannerIdx > 0 {
		lastBeforeResumeSec = frames[resumeBannerIdx-1].sec
	}

	// Find the first frame AFTER the banner that contains actual resumed session data
	// Skip the banner frame itself and any empty frames
	var firstAfterResumeSec uint32 = 0
	var firstAfterResumeIdx int = -1
	for i := resumeBannerIdx + 1; i < len(frames); i++ {
		if len(frames[i].payload) > 0 {
			firstAfterResumeSec = frames[i].sec
			firstAfterResumeIdx = i
			break
		}
	}

	// No actual resumed data found after the banner
	if firstAfterResumeIdx < 0 || firstAfterResumeSec == 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG] No resumed data found after banner. firstAfterResumeIdx=%d, firstAfterResumeSec=%d\n",
			firstAfterResumeIdx, firstAfterResumeSec)
		return nil
	}

	// Calculate the time gap between the original session and the resumed session
	timeDelta := int64(firstAfterResumeSec) - int64(lastBeforeResumeSec)

	// If the gap is small (<= 5 seconds), no need to normalize
	if timeDelta <= 5 {
		return nil
	}

	// We want a small pause (3 seconds) at the resume point
	// The first resumed frame should be at lastBeforeResumeSec + 3
	// So we subtract (firstAfterResumeSec - (lastBeforeResumeSec + 3)) from all resumed frames
	targetFirstResumedSec := lastBeforeResumeSec + 3
	adjustment := int64(firstAfterResumeSec) - int64(targetFirstResumedSec)

	// If adjustment would be negative, don't adjust
	if adjustment <= 0 {
		return nil
	}

	var buf bytes.Buffer
	var lastWrittenSec uint32 = 0
	var lastWrittenUsec uint32 = 0

	for i, f := range frames {
		// Skip the banner frame entirely
		if i == resumeBannerIdx {
			continue
		}

		sec := f.sec
		usec := f.usec

		// Apply adjustment to frames after the banner
		if i > resumeBannerIdx {
			// Calculate new timestamp by subtracting the adjustment
			newSec := int64(sec) - adjustment
			if newSec < 0 {
				newSec = 0
			}
			sec = uint32(newSec)

			// Ensure monotonic timestamps - each frame must be >= previous written frame
			if sec < lastWrittenSec || (sec == lastWrittenSec && usec <= lastWrittenUsec) {
				sec = lastWrittenSec
				usec = lastWrittenUsec + 1
				if usec >= 1000000 {
					sec++
					usec -= 1000000
				}
			}
		}

		writeRecord(&buf, sec, usec, f.payload)
		lastWrittenSec = sec
		lastWrittenUsec = usec
	}

	return os.WriteFile(ttyPath, buf.Bytes(), 0644)
}
