package index

import "fmt"

// Index represents a column index for fast lookups
type Index struct {
	ColumnName string              // Column being indexed
	SourceHash string              // SHA256 hash of source file
	HashMap    map[string][]int    // Value -> row positions
	CreatedAt  int64              // Creation timestamp
	Valid      bool               // Whether index is still valid
}

// NewIndex creates a new index
func NewIndex(columnName, sourceHash string) *Index {
	return &Index{
		ColumnName: columnName,
		SourceHash: sourceHash,
		HashMap:    make(map[string][]int),
		Valid:      true,
	}
}

// Lookup returns row positions for a given value
func (i *Index) Lookup(value string) []int {
	if positions, ok := i.HashMap[value]; ok {
		return positions
	}
	return []int{}
}

// IsValid checks if index is still valid
func (i *Index) IsValid() bool {
	return i.Valid
}

// SourceHashValue returns the source file hash
func (i *Index) SourceHashValue() string {
	return i.SourceHash
}

// Add adds a row position for a value
func (i *Index) Add(value string, rowPosition int) {
	if i.HashMap[value] == nil {
		i.HashMap[value] = []int{}
	}
	i.HashMap[value] = append(i.HashMap[value], rowPosition)
}

// String returns string representation
func (i *Index) String() string {
	return fmt.Sprintf("Index(%s, hash=%s, valid=%v, entries=%d)",
		i.ColumnName, i.SourceHash, i.Valid, len(i.HashMap))
}
