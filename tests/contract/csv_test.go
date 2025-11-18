package contract

import (
	"bytes"
	"testing"

	"github.com/yourusername/yatisql/internal/csv"
)

// TestCSVParsingQuotedFields tests RFC 4180 CSV parsing with quoted fields
func TestCSVParsingQuotedFields(t *testing.T) {
	input := `"hello","world"`
	expected := []string{"hello", "world"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingEmbeddedNewlines tests handling of newlines within quoted fields
func TestCSVParsingEmbeddedNewlines(t *testing.T) {
	input := `"line1
line2","field2"`
	expected := []string{"line1\nline2", "field2"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingEscapedQuotes tests handling of escaped quotes within quoted fields
func TestCSVParsingEscapedQuotes(t *testing.T) {
	input := `"hello ""world""","test"`
	expected := []string{`hello "world"`, "test"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingUnquotedFields tests handling of unquoted fields
func TestCSVParsingUnquotedFields(t *testing.T) {
	input := `hello,world,test`
	expected := []string{"hello", "world", "test"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingEmptyFields tests handling of empty fields
func TestCSVParsingEmptyFields(t *testing.T) {
	input := `hello,,world`
	expected := []string{"hello", "", "world"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingMixedQuoting tests RFC 4180 with mixed quoted and unquoted fields
func TestCSVParsingMixedQuoting(t *testing.T) {
	input := `hello,"world",test,"another field"`
	expected := []string{"hello", "world", "test", "another field"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingWithCommaInQuotedField tests commas within quoted fields
func TestCSVParsingWithCommaInQuotedField(t *testing.T) {
	input := `"hello, world","test,value"`
	expected := []string{"hello, world", "test,value"}

	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(result))
	}
	for i, field := range result {
		if field != expected[i] {
			t.Errorf("Field %d: expected %q, got %q", i, expected[i], field)
		}
	}
}

// TestCSVParsingEdgeCases tests various edge cases
func TestCSVParsingEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single field", "hello", []string{"hello"}},
		{"quoted single field", `"hello"`, []string{"hello"}},
		{"trailing comma", `hello,world,`, []string{"hello", "world", ""}},
		{"leading comma", `,hello,world`, []string{"", "hello", "world"}},
		{"quoted empty field", `hello,"",world`, []string{"hello", "", "world"}},
		{"only quotes", `""`, []string{""}},
		{"field with spaces", `hello  ,  world`, []string{"hello  ", "  world"}},
		{"quoted field with spaces", `"hello  ","  world"`, []string{"hello  ", "  world"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := csv.ParseCSVRow(tc.input)
			if err != nil {
				t.Fatalf("ParseCSVRow failed: %v", err)
			}
			if len(result) != len(tc.expected) {
				t.Fatalf("Expected %d fields, got %d", len(tc.expected), len(result))
			}
			for i, field := range result {
				if field != tc.expected[i] {
					t.Errorf("Field %d: expected %q, got %q", i, tc.expected[i], field)
				}
			}
		})
	}
}

// TestCSVStreamRowIterator tests streaming row iteration without loading entire file
func TestCSVStreamRowIterator(t *testing.T) {
	csvData := `id,name,age
1,Alice,28
2,Bob,35
3,Charlie,22`

	expected := [][]string{
		{"id", "name", "age"},
		{"1", "Alice", "28"},
		{"2", "Bob", "35"},
		{"3", "Charlie", "22"},
	}

	reader := bytes.NewReader([]byte(csvData))
	iterator, err := csv.NewRowIterator(reader, ',')
	if err != nil {
		t.Fatalf("NewRowIterator failed: %v", err)
	}
	defer iterator.Close()

	rowIdx := 0
	for iterator.Next() {
		row := iterator.Row()
		if rowIdx >= len(expected) {
			t.Fatalf("Unexpected row %d", rowIdx)
		}
		if len(row) != len(expected[rowIdx]) {
			t.Fatalf("Row %d: expected %d fields, got %d", rowIdx, len(expected[rowIdx]), len(row))
		}
		for colIdx, field := range row {
			if field != expected[rowIdx][colIdx] {
				t.Errorf("Row %d, Field %d: expected %q, got %q", rowIdx, colIdx, expected[rowIdx][colIdx], field)
			}
		}
		rowIdx++
	}

	if rowIdx != len(expected) {
		t.Errorf("Expected %d rows, got %d", len(expected), rowIdx)
	}
}

// TestCSVWriterPreserveFormat tests that CSV writer preserves original format
func TestCSVWriterPreserveFormat(t *testing.T) {
	originalRow := []string{"hello", `world "quoted"`, "test,comma"}
	delimiter := ','

	// Write row
	output, err := csv.WriteCSVRow(originalRow, delimiter)
	if err != nil {
		t.Fatalf("WriteCSVRow failed: %v", err)
	}

	// Parse it back
	parsed, err := csv.ParseCSVRow(output)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}

	// Verify round-trip
	if len(parsed) != len(originalRow) {
		t.Fatalf("Expected %d fields, got %d", len(originalRow), len(parsed))
	}
	for i, field := range parsed {
		if field != originalRow[i] {
			t.Errorf("Field %d round-trip failed: expected %q, got %q", i, originalRow[i], field)
		}
	}
}

// TestCSVRoundTrip tests that CSV data can be written and read back
func TestCSVRoundTrip(t *testing.T) {
	// Test round-trip for complex CSV data
	input := `name,value,"quoted,field","with""quotes"`
	result, err := csv.ParseCSVRow(input)
	if err != nil {
		t.Fatalf("ParseCSVRow failed: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(result))
	}
	if result[2] != "quoted,field" {
		t.Errorf("Expected 'quoted,field', got %q", result[2])
	}
}

