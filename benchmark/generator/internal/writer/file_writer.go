package writer

import (
	"path/filepath"
)

// FileWriter manages output file creation for different data formats
type FileWriter struct {
	OutputDir string
}

func NewFileWriter(dir string) *FileWriter {
	return &FileWriter{
		OutputDir: dir,
	}
}

// CSVPath returns full path for a CSV file
func (f *FileWriter) CSVPath(filename string) string {
	return filepath.Join(f.OutputDir, filename+".csv")
}

// NDJSONPath returns full path for an NDJSON file
func (f *FileWriter) NDJSONPath(filename string) string {
	return filepath.Join(f.OutputDir, filename+".ndjson")
}

// Close is a no-op for FileWriter (individual writers manage their own lifecycle)
func (f *FileWriter) Close() {
	// Individual writers are closed by their creators
}
