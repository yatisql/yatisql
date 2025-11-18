package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// ParseCSVRow parses a single CSV row according to RFC 4180
// Handles:
// - Quoted fields
// - Escaped quotes (double quotes within quoted fields)
// - Newlines within quoted fields
// - Empty fields
// - Mixed quoted and unquoted fields
func ParseCSVRow(row string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(row))
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	record, err := reader.Read()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse CSV row: %w", err)
	}
	return record, nil
}

// CSVRowParser handles streaming CSV parsing with proper RFC 4180 compliance
type CSVRowParser struct {
	reader *csv.Reader
}

// NewCSVRowParser creates a parser from an io.Reader
func NewCSVRowParser(r io.Reader, delimiter rune) *CSVRowParser {
	csvReader := csv.NewReader(r)
	csvReader.Comma = delimiter
	csvReader.FieldsPerRecord = -1 // Allow variable fields (some rows may have different column counts)
	return &CSVRowParser{
		reader: csvReader,
	}
}

// ParseRow reads and parses next row
func (p *CSVRowParser) ParseRow() ([]string, error) {
	return p.reader.Read()
}

// All returns all remaining rows
func (p *CSVRowParser) All() ([][]string, error) {
	return p.reader.ReadAll()
}
