package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
)

// CSVWriter handles writing CSV rows to file
type CSVWriter struct {
	file       *os.File
	writer     *csv.Writer
	delimiter  rune
	mu         sync.Mutex
}

// NewCSVWriter creates a new CSV writer for file
func NewCSVWriter(path string, delimiter rune) (*CSVWriter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &CSVWriter{
		file:      file,
		writer:    csv.NewWriter(file),
		delimiter: delimiter,
	}, nil
}

// WriteRow writes a single row in CSV format
// Properly escapes quoted fields and preserves format
func WriteCSVRow(fields []string, delimiter rune) (string, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	writer.Comma = delimiter

	err := writer.Write(fields)
	if err != nil {
		return "", fmt.Errorf("failed to write CSV row: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	// Remove trailing newline that csv.Writer adds
	result := buf.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

// WriteRows writes multiple rows to buffer
func (w *CSVWriter) WriteRows(rows [][]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, row := range rows {
		err := w.writer.Write(row)
		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	w.writer.Flush()
	return w.writer.Error()
}

// AppendRows appends rows to file
func (w *CSVWriter) AppendRows(rows [][]string) error {
	return w.WriteRows(rows)
}

// UpdateRow updates a row at specific position (placeholder - full impl in next phase)
func (w *CSVWriter) UpdateRow(index int, row []string) error {
	// This will be implemented with proper file rewriting in Phase 3
	return fmt.Errorf("UpdateRow not yet implemented")
}

// Flush writes buffered data to file
func (w *CSVWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writer.Flush()
	return w.writer.Error()
}

// Close closes the writer
func (w *CSVWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return err
	}
	return w.file.Close()
}

// WriteBuffer writes rows to buffer
func WriteBuffer(w io.Writer, rows [][]string, delimiter rune) error {
	writer := csv.NewWriter(w)
	writer.Comma = delimiter

	for _, row := range rows {
		err := writer.Write(row)
		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}
