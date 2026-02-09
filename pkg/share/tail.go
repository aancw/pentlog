package share

import (
	"context"
	"io"
	"os"
	"time"
)

const (
	tailPollInterval = 100 * time.Millisecond
	tailReadBufSize  = 64 * 1024
)

type Tailer struct {
	path      string
	fromStart bool
}

func NewTailer(path string, fromStart bool) *Tailer {
	return &Tailer{
		path:      path,
		fromStart: fromStart,
	}
}

func (t *Tailer) Run(ctx context.Context, out chan<- []byte) error {
	f, err := t.waitForFile(ctx)
	if err != nil {
		return err
	}
	defer f.Close()

	var offset int64
	if !t.fromStart {
		info, err := f.Stat()
		if err != nil {
			return err
		}
		offset = info.Size()
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return err
		}
	}

	buf := make([]byte, tailReadBufSize)
	ticker := time.NewTicker(tailPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := os.Stat(t.path)
			if err != nil {
				continue
			}

			currentSize := info.Size()
			if currentSize < offset {
				offset = 0
				f.Seek(0, io.SeekStart)
			}

			if currentSize == offset {
				continue
			}

			for {
				n, err := f.Read(buf)
				if n > 0 {
					data := make([]byte, n)
					copy(data, buf[:n])
					offset += int64(n)

					select {
					case out <- data:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				if err == io.EOF || n == 0 {
					break
				}
				if err != nil {
					return err
				}
			}
		}
	}
}

func (t *Tailer) waitForFile(ctx context.Context) (*os.File, error) {
	ticker := time.NewTicker(tailPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			f, err := os.Open(t.path)
			if err == nil {
				return f, nil
			}
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}
}
