package index

import (
	"context"
	"fmt"
	"sync"
)

// IndexManager manages index creation, validation, and caching
type IndexManager struct {
	indexes map[string]*Index // Key: "filename:column"
	mu      sync.RWMutex
}

// NewIndexManager creates a new index manager
func NewIndexManager() *IndexManager {
	return &IndexManager{
		indexes: make(map[string]*Index),
	}
}

// GetOrCreateIndex returns existing index or creates new one
func (m *IndexManager) GetOrCreateIndex(ctx context.Context, file CSVFileInterface, column string) (*Index, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", file.GetPath(), column)

	// Check if we have a cached index
	if idx, ok := m.indexes[key]; ok {
		if m.isValidIndex(idx, file) {
			return idx, nil
		}
		// Index is invalid, rebuild it
		newIdx, err := m.buildIndex(file, column)
		if err != nil {
			return nil, err
		}
		m.indexes[key] = newIdx
		return newIdx, nil
	}

	// Create new index
	newIdx, err := m.buildIndex(file, column)
	if err != nil {
		return nil, err
	}

	m.indexes[key] = newIdx
	return newIdx, nil
}

// IsValid checks if index is still valid for file
func (m *IndexManager) IsValid(index *Index, file CSVFileInterface) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isValidIndex(index, file)
}

// isValidIndex is the internal version without locking
func (m *IndexManager) isValidIndex(index *Index, file CSVFileInterface) bool {
	if index == nil {
		return false
	}

	currentHash, _ := file.ComputeHash()
	return index.SourceHash == currentHash
}

// RebuildIndex recreates index for file
func (m *IndexManager) RebuildIndex(ctx context.Context, file CSVFileInterface, column string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	newIdx, err := m.buildIndex(file, column)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", file.GetPath(), column)
	m.indexes[key] = newIdx
	return nil
}

// DeleteIndex removes index from storage
func (m *IndexManager) DeleteIndex(file CSVFileInterface, column string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", file.GetPath(), column)
	delete(m.indexes, key)
	return nil
}

// buildIndex creates a new index from file data
func (m *IndexManager) buildIndex(file CSVFileInterface, column string) (*Index, error) {
	hash, err := file.ComputeHash()
	if err != nil {
		return nil, fmt.Errorf("failed to compute file hash: %w", err)
	}

	idx := NewIndex(column, hash)

	// If file has a precomputed row index map, use it
	if rowMap := file.GetRowIndexMap(); rowMap != nil {
		for value, positions := range rowMap {
			for _, pos := range positions {
				idx.Add(value, pos)
			}
		}
	}

	return idx, nil
}

// CSVFileInterface defines the interface for CSV files (for testing)
type CSVFileInterface interface {
	ComputeHash() (string, error)
	GetPath() string
	GetRowIndexMap() map[string][]int
}
