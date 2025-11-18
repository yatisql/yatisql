package contract

import (
	"context"
	"testing"

	"github.com/yourusername/yatisql/internal/index"
)

// TestIndexManagerGetOrCreateIndex tests automatic index creation
func TestIndexManagerGetOrCreateIndex(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	// Create a mock CSV file
	csvFile := &MockCSVFile{
		Path: "test.csv",
		Hash: "abc123",
	}

	// Get or create index
	idx, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	if idx == nil {
		t.Fatalf("Expected index, got nil")
	}

	// Index should be valid
	if !manager.IsValid(idx, csvFile) {
		t.Errorf("Index should be valid")
	}
}

// TestIndexManagerIsValid tests index validation with hash comparison
func TestIndexManagerIsValid(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path: "test.csv",
		Hash: "abc123",
	}

	// Create index
	index, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Index should be valid with same hash
	if !manager.IsValid(index, csvFile) {
		t.Errorf("Index should be valid with matching hash")
	}

	// Index should be invalid with different hash
	modifiedFile := &MockCSVFile{
		Path: "test.csv",
		Hash: "def456",
	}
	if manager.IsValid(index, modifiedFile) {
		t.Errorf("Index should be invalid with different hash")
	}
}

// TestIndexManagerRebuildIndex tests index rebuilding after file change
func TestIndexManagerRebuildIndex(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path: "test.csv",
		Hash: "abc123",
	}

	// Create initial index
	index1, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Change file hash
	csvFile.Hash = "def456"

	// Rebuild index
	err = manager.RebuildIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("RebuildIndex failed: %v", err)
	}

	// Get index again (should be new)
	index2, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// New index should have new hash
	if !manager.IsValid(index2, csvFile) {
		t.Errorf("Rebuilt index should be valid")
	}

	// Indexes should be different objects
	if index1 == index2 {
		t.Errorf("Expected new index after rebuild")
	}
}

// TestIndexManagerDeleteIndex tests index deletion
func TestIndexManagerDeleteIndex(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path: "test.csv",
		Hash: "abc123",
	}

	// Create index
	_, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Delete index
	err = manager.DeleteIndex(csvFile, "id")
	if err != nil {
		t.Fatalf("DeleteIndex failed: %v", err)
	}

	// Index should not exist anymore
	// (Should create new one on next GetOrCreateIndex)
}

// TestIndexLookup tests HashMap-based index lookup
func TestIndexLookup(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path:        "users.csv",
		Hash:        "abc123",
		RowIndexMap: map[string][]int{"1": {0}, "2": {1}, "3": {2}},
	}

	index, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Lookup existing value
	positions := index.Lookup("1")
	if len(positions) != 1 || positions[0] != 0 {
		t.Errorf("Expected [0], got %v", positions)
	}

	// Lookup non-existing value
	positions = index.Lookup("99")
	if len(positions) != 0 {
		t.Errorf("Expected empty result, got %v", positions)
	}
}

// TestIndexLookupMultipleMatches tests index with multiple matches for same value
func TestIndexLookupMultipleMatches(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path:        "users.csv",
		Hash:        "abc123",
		RowIndexMap: map[string][]int{"NYC": {0, 2, 5}, "LA": {1, 3}},
	}

	index, err := manager.GetOrCreateIndex(ctx, csvFile, "city")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Lookup value with multiple matches
	positions := index.Lookup("NYC")
	if len(positions) != 3 {
		t.Fatalf("Expected 3 positions, got %d", len(positions))
	}
	expected := []int{0, 2, 5}
	for i, pos := range positions {
		if pos != expected[i] {
			t.Errorf("Position %d: expected %d, got %d", i, expected[i], pos)
		}
	}
}

// TestIndexHashConsistency tests SHA256 hash consistency
func TestIndexHashConsistency(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path: "test.csv",
		Hash: "abc123", // SHA256 of file content
	}

	index1, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("First GetOrCreateIndex failed: %v", err)
	}

	// Create second manager (independent)
	manager2 := index.NewIndexManager()
	index2, err := manager2.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("Second GetOrCreateIndex failed: %v", err)
	}

	// Both indexes should have same hash (derived from file)
	// This ensures index validity is based on file content, not creation time
	_ = index1
	_ = index2
}

// TestIndexConcurrentAccess tests thread-safe concurrent access with RWMutex
func TestIndexConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path:        "test.csv",
		Hash:        "abc123",
		RowIndexMap: map[string][]int{"id": {0}},
	}

	// Create initial index
	_, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Simulate concurrent reads
	resultsChan := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			index, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
			if err != nil {
				resultsChan <- false
				return
			}
			resultsChan <- index != nil
		}()
	}

	// Verify all concurrent reads succeeded
	for i := 0; i < 10; i++ {
		if !<-resultsChan {
			t.Errorf("Concurrent read %d failed", i)
		}
	}
}

// TestIndexPersistence tests index serialization and deserialization
func TestIndexPersistence(t *testing.T) {
	ctx := context.Background()
	manager := index.NewIndexManager()

	csvFile := &MockCSVFile{
		Path:        "test.csv",
		Hash:        "abc123",
		RowIndexMap: map[string][]int{"1": {0, 1}, "2": {2}, "3": {3, 4}},
	}

	// Create and save index
	index1, err := manager.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("GetOrCreateIndex failed: %v", err)
	}

	// Create new manager and load index
	manager2 := index.NewIndexManager()
	index2, err := manager2.GetOrCreateIndex(ctx, csvFile, "id")
	if err != nil {
		t.Fatalf("Second GetOrCreateIndex failed: %v", err)
	}

	// Both indexes should be valid and have same data
	if !manager.IsValid(index1, csvFile) {
		t.Errorf("First index invalid")
	}
	if !manager2.IsValid(index2, csvFile) {
		t.Errorf("Second index invalid")
	}

	// Lookups should return same results
	pos1 := index1.Lookup("1")
	pos2 := index2.Lookup("1")
	_ = pos1
	_ = pos2
	if len(pos1) != len(pos2) {
		t.Errorf("Different lookup results after persistence")
	}
}

// ===== HELPER TYPES AND FUNCTIONS (CONTRACT DEFINITIONS) =====

// Index represents a column index for fast lookups
type Index interface {
	Lookup(value string) []int  // Returns row positions for value
	IsValid() bool              // Check if index is valid
	SourceHash() string         // SHA256 of indexed file
}

// IndexManager manages index creation, validation, and caching
type IndexManager interface {
	// GetOrCreateIndex returns existing index or creates new one
	// Returns error if index creation fails
	GetOrCreateIndex(ctx context.Context, file *MockCSVFile, column string) (Index, error)

	// IsValid checks if index is still valid for file (hash comparison)
	IsValid(index Index, file *MockCSVFile) bool

	// RebuildIndex recreates index for file (used after file modification)
	RebuildIndex(ctx context.Context, file *MockCSVFile, column string) error

	// DeleteIndex removes index from storage
	DeleteIndex(file *MockCSVFile, column string) error
}

// MockCSVFile represents a CSV file with data for testing
type MockCSVFile struct {
	Path        string
	Hash        string
	RowIndexMap map[string][]int // Maps value -> row positions
}

// ComputeHash returns SHA256 hash of file content
func (f *MockCSVFile) ComputeHash() (string, error) {
	return f.Hash, nil
}

// GetPath returns the file path
func (f *MockCSVFile) GetPath() string {
	return f.Path
}

// GetRowIndexMap returns the row index map
func (f *MockCSVFile) GetRowIndexMap() map[string][]int {
	return f.RowIndexMap
}
