package logs

import (
	"bytes"
	"errors"
	"os"
	"sync"
)

type rotatingWriter struct {
	mu       sync.Mutex
	path     string
	maxLines int
	lines    [][]byte
}

func newRotatingWriter(path string, maxLines int) (*rotatingWriter, error) {
	w := &rotatingWriter{path: path, maxLines: maxLines}

	if err := w.loadExisting(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *rotatingWriter) loadExisting() error {
	data, err := os.ReadFile(w.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	lines := bytes.Split(bytes.TrimRight(data, "\n"), []byte("\n"))
	if len(lines) == 1 && len(lines[0]) == 0 {
		return nil
	}

	if len(lines) > w.maxLines {
		lines = lines[len(lines)-w.maxLines:]
	}

	w.lines = lines
	return nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	line := append([]byte(nil), bytes.TrimRight(p, "\n")...)
	w.lines = append(w.lines, line)
	if len(w.lines) > w.maxLines {
		w.lines = w.lines[len(w.lines)-w.maxLines:]
	}

	if err := w.flush(); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *rotatingWriter) flush() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, line := range w.lines {
		if _, err := f.Write(line); err != nil {
			return err
		}
		if _, err := f.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}
