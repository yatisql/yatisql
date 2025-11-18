package csv

import (
	"fmt"
	"io"
)

// StreamingRowReader implements RowIterator for streaming CSV rows
type StreamingRowReader struct {
	parser *CSVRowParser
	file   io.Closer
	row    []string
	err    error
	eof    bool
}

// NewRowIterator creates a row iterator for streaming CSV data
// Handles:
// - Multi-line quoted fields
// - Proper delimiter detection
// - Efficient streaming without loading entire file
func NewRowIterator(reader interface{}, delimiter rune) (RowIterator, error) {
	var r io.Reader
	var closer io.Closer

	// Handle different input types
	switch v := reader.(type) {
	case io.Reader:
		r = v
		if c, ok := v.(io.Closer); ok {
			closer = c
		}
	default:
		return nil, fmt.Errorf("unsupported reader type")
	}

	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}

	parser := NewCSVRowParser(r, delimiter)
	return &StreamingRowReader{
		parser: parser,
		file:   closer,
		eof:    false,
	}, nil
}

// Next advances to next row and returns true if row exists
func (r *StreamingRowReader) Next() bool {
	if r.eof {
		return false
	}

	row, err := r.parser.ParseRow()
	if err != nil && err != io.EOF {
		r.err = err
		return false
	}

	if err == io.EOF {
		r.eof = true
		r.row = row
		return len(row) > 0 && row[0] != "" // Return true if last row has data
	}

	r.row = row
	return true
}

// Row returns current row as string slice
func (r *StreamingRowReader) Row() []string {
	return r.row
}

// Close closes the reader
func (r *StreamingRowReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Err returns any error encountered during reading
func (r *StreamingRowReader) Err() error {
	return r.err
}
