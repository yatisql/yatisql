package csv

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// CSVFile represents a CSV or TSV file with metadata
type CSVFile struct {
	Path       string // File path (e.g., "data.csv" or "users.csv.gz")
	Delimiter  rune   // ',' for CSV or '\t' for TSV
	Encoding   string // "utf-8" (default) or other encodings
	HasHeader  bool   // Whether first row is header
	Headers    []string // Column names from header row
	SourceHash string    // SHA256 hash of file content for index validation
	mu         sync.RWMutex // Protects concurrent access
}

// OpenFile opens a CSV/TSV file and auto-detects format
func OpenFile(path string) (*CSVFile, error) {
	// Detect format from extension
	delimiter := detectDelimiter(path)

	// Read headers from file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first row for headers
	reader := NewPlainRowReader(file, delimiter)
	headers := []string{}
	if reader.Next() {
		headers = reader.Row()
	}

	// Compute file hash
	file, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hash, err := computeHash(file)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}

	return &CSVFile{
		Path:       path,
		Delimiter:  delimiter,
		Encoding:   "utf-8",
		HasHeader:  true,
		Headers:    headers,
		SourceHash: hash,
	}, nil
}

// detectDelimiter determines if file is CSV or TSV based on extension
func detectDelimiter(path string) rune {
	if strings.HasSuffix(strings.ToLower(path), ".tsv") {
		return '\t'
	}
	// Handle .tsv.gz
	if strings.HasSuffix(strings.ToLower(path), ".tsv.gz") {
		return '\t'
	}
	return ','
}

// ComputeHash returns SHA256 hash of file content
func (f *CSVFile) ComputeHash() (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, err := os.Open(f.Path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return computeHash(file)
}

// computeHash computes SHA256 hash of file content
func computeHash(file *os.File) (string, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// StreamRows returns a row iterator for streaming the file
func (f *CSVFile) StreamRows() (RowIterator, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, err := os.Open(f.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return NewRowIterator(file, f.Delimiter)
}

// RowIterator interface for streaming row reading
type RowIterator interface {
	Next() bool
	Row() []string
	Close() error
}

// plainRowReader implements RowIterator for uncompressed files
type plainRowReader struct {
	file   *os.File
	reader io.Reader
	lines  []string
	index  int
}

// NewPlainRowReader creates a simple row reader for uncompressed files
func NewPlainRowReader(file io.Reader, delimiter rune) *plainRowReader {
	return &plainRowReader{
		reader: file,
		lines:  []string{},
		index:  0,
	}
}

// Next reads next row (placeholder - will be properly implemented in T010)
func (r *plainRowReader) Next() bool {
	return false
}

// Row returns current row
func (r *plainRowReader) Row() []string {
	return []string{}
}

// Close closes the reader
func (r *plainRowReader) Close() error {
	if rf, ok := r.reader.(*os.File); ok {
		return rf.Close()
	}
	return nil
}
