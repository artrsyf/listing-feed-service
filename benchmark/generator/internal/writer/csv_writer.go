package writer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CSVWriter is high-performance streaming writer for PostgreSQL COPY
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
	buf    *bufio.Writer

	mu sync.Mutex
}

func NewCSVWriter(dir string, filename string) (*CSVWriter, error) {
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriterSize(f, 4*1024*1024) // 4MB buffer

	w := csv.NewWriter(buf)
	w.Comma = ','

	return &CSVWriter{
		file:   f,
		writer: w,
		buf:    buf,
	}, nil
}

// WriteRow writes single CSV row
func (w *CSVWriter) WriteRow(row []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.writer.Write(row)
}

// WriteBatch writes multiple rows (faster path)
func (w *CSVWriter) WriteBatch(rows [][]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, r := range rows {
		if err := w.writer.Write(r); err != nil {
			return err
		}
	}

	return nil
}

// Flush forces buffer flush
func (w *CSVWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writer.Flush()
	return w.buf.Flush()
}

// Close finalizes file
func (w *CSVWriter) Close() error {
	w.Flush()

	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}
