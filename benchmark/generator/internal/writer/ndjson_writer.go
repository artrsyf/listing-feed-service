package writer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type NDJSONWriter struct {
	file *os.File
	buf  *bufio.Writer
}

func NewNDJSONWriter(dir string, filename string) (*NDJSONWriter, error) {
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriterSize(f, 4*1024*1024)

	return &NDJSONWriter{
		file: f,
		buf:  buf,
	}, nil
}

// WriteBulkItem writes Elasticsearch bulk format:
// { "index": {} }
// { document }
func (w *NDJSONWriter) WriteBulkItem(action string, doc string) error {
	if _, err := fmt.Fprintf(w.buf, `{"%s":{}}`+"\n", action); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w.buf, doc); err != nil {
		return err
	}

	return nil
}

func (w *NDJSONWriter) Flush() error {
	return w.buf.Flush()
}

func (w *NDJSONWriter) Close() error {
	w.Flush()
	return w.file.Close()
}
