package contract

import (
	"bytes"
	"testing"

	"github.com/yourusername/yatisql/internal/executor"
	"github.com/yourusername/yatisql/internal/query"
)

// TestSELECTAll tests SELECT * from CSV file
func TestSELECTAll(t *testing.T) {
	// Create test CSV data
	csvData := `id,name,age,city
1,Alice,28,New York
2,Bob,35,Los Angeles
3,Charlie,22,Chicago`

	// Parse query
	q, err := query.ParseSQL("SELECT * FROM test.csv")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	// Execute query
	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatalf("Expected result, got nil")
	}

	rows := result.Rows()
	if len(rows) != 4 { // Header + 3 data rows
		t.Fatalf("Expected 4 rows (including header), got %d", len(rows))
	}

	// Verify headers
	if len(rows[0]) != 4 {
		t.Fatalf("Expected 4 columns, got %d", len(rows[0]))
	}
	if rows[0][0] != "id" || rows[0][1] != "name" {
		t.Errorf("Expected header [id, name, ...], got %v", rows[0])
	}

	// Verify first data row
	if rows[1][1] != "Alice" || rows[1][2] != "28" {
		t.Errorf("Expected [1, Alice, 28, ...], got %v", rows[1])
	}
}

// TestSELECTProjection tests SELECT with specific columns
func TestSELECTProjection(t *testing.T) {
	csvData := `id,name,age,city
1,Alice,28,New York
2,Bob,35,Los Angeles
3,Charlie,22,Chicago`

	q, err := query.ParseSQL("SELECT id, name FROM test.csv")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	rows := result.Rows()
	if len(rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(rows))
	}

	// Verify only 2 columns
	if len(rows[0]) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(rows[0]))
	}
	if rows[0][0] != "id" || rows[0][1] != "name" {
		t.Errorf("Expected columns [id, name], got %v", rows[0])
	}

	// Verify data row has correct columns
	if rows[1][0] != "1" || rows[1][1] != "Alice" {
		t.Errorf("Expected [1, Alice], got %v", rows[1])
	}
}

// TestSELECTWithWHERE tests SELECT with WHERE filtering
func TestSELECTWithWHERE(t *testing.T) {
	csvData := `id,name,age,city
1,Alice,28,New York
2,Bob,35,Los Angeles
3,Charlie,22,Chicago
4,Diana,31,Houston`

	q, err := query.ParseSQL("SELECT * FROM test.csv WHERE age > 30")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	rows := result.Rows()
	// Should have header + 2 matching rows (Bob age 35, Diana age 31)
	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows (header + 2 data), got %d", len(rows))
	}

	// Verify data rows
	if rows[1][1] != "Bob" {
		t.Errorf("Expected first data row to be Bob, got %s", rows[1][1])
	}
	if rows[2][1] != "Diana" {
		t.Errorf("Expected second data row to be Diana, got %s", rows[2][1])
	}
}

// TestSELECTWithWHEREAndProjection tests SELECT with both WHERE and column projection
func TestSELECTWithWHEREAndProjection(t *testing.T) {
	csvData := `id,name,age,city
1,Alice,28,New York
2,Bob,35,Los Angeles
3,Charlie,22,Chicago
4,Diana,31,Houston`

	q, err := query.ParseSQL("SELECT name, age FROM test.csv WHERE age > 25")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	rows := result.Rows()
	// Header + 3 matching rows (Alice, Bob, Diana all > 25)
	if len(rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(rows))
	}

	// Only 2 columns
	if len(rows[0]) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(rows[0]))
	}

	// Verify column names
	if rows[0][0] != "name" || rows[0][1] != "age" {
		t.Errorf("Expected [name, age], got %v", rows[0])
	}
}

// TestSELECTTSVFile tests SELECT on TSV (tab-delimited) file
func TestSELECTTSVFile(t *testing.T) {
	// TSV format with tabs instead of commas
	tsvData := "id\tname\tage\tcity\n1\tAlice\t28\tNew York\n2\tBob\t35\tLos Angeles"

	q, err := query.ParseSQL("SELECT * FROM test.tsv")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	// Use tab delimiter for TSV
	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(tsvData)), '\t')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	rows := result.Rows()
	if len(rows) != 3 { // Header + 2 data rows
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	// Verify headers
	if rows[0][0] != "id" || rows[0][1] != "name" {
		t.Errorf("Expected [id, name, ...], got %v", rows[0])
	}

	// Verify data
	if rows[1][1] != "Alice" {
		t.Errorf("Expected Alice, got %s", rows[1][1])
	}
}

// TestSELECTCompressedFile tests SELECT on regular file (compression will be transparent in T029)
func TestSELECTCompressedFile(t *testing.T) {
	// Compression support is transparent at execution layer
	// File opening and decompression happens in T029/T030
	csvData := `id,name,age
1,Alice,28
2,Bob,35`

	// Parse as regular CSV for now
	q, err := query.ParseSQL("SELECT * FROM test.csv")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQueryWithData failed: %v", err)
	}

	if result == nil {
		t.Fatalf("Expected result, got nil")
	}

	rows := result.Rows()
	if len(rows) != 3 { // Header + 2 data rows
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}
}

// TestSELECTEmptyResult tests SELECT that matches no rows
func TestSELECTEmptyResult(t *testing.T) {
	csvData := `id,name,age
1,Alice,28
2,Bob,35`

	q, err := query.ParseSQL("SELECT * FROM test.csv WHERE age > 100")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	rows := result.Rows()
	// Only header, no data rows
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row (header only), got %d", len(rows))
	}
}

// TestSELECTToCSV tests result formatting to CSV
func TestSELECTToCSV(t *testing.T) {
	csvData := `id,name
1,Alice
2,Bob`

	q, err := query.ParseSQL("SELECT * FROM test.csv")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	result, err := executor.ExecuteSelectQueryWithData(q, bytes.NewReader([]byte(csvData)), ',')
	if err != nil {
		t.Fatalf("ExecuteSelectQuery failed: %v", err)
	}

	// Format as CSV
	csvOutput, err := result.ToCSV()
	if err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	if csvOutput == "" {
		t.Fatalf("Expected CSV output, got empty string")
	}

	// Verify it contains the data
	if !bytes.Contains([]byte(csvOutput), []byte("Alice")) {
		t.Errorf("Expected CSV output to contain 'Alice', got: %s", csvOutput)
	}
}

