package executor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yourusername/yatisql/internal/csv"
	"github.com/yourusername/yatisql/internal/eval"
	"github.com/yourusername/yatisql/internal/query"
)

// ExecuteSelectQuery executes a SELECT query
func ExecuteSelectQuery(q query.Query) (Result, error) {
	if q == nil {
		return nil, fmt.Errorf("query is nil")
	}

	switch q.Type() {
	case query.QueryTypeSelect:
		return executeSelect(q.(*query.SelectQuery))
	case query.QueryTypeJoin:
		return executeSelect(q.(*query.JoinQuery))
	default:
		return nil, fmt.Errorf("unsupported query type for SELECT: %v", q.Type())
	}
}

// executeSelect executes a SELECT query on CSV data
func executeSelect(q interface{}) (Result, error) {
	var columns []string
	var tables []string
	var whereExpr query.Expression

	// Handle both SelectQuery and JoinQuery
	switch q := q.(type) {
	case *query.SelectQuery:
		columns = q.Columns()
		tables = q.Tables()
		whereExpr = q.WhereClause()
	case *query.JoinQuery:
		columns = q.Columns()
		tables = q.Tables()
		whereExpr = q.WhereClause()
	default:
		return nil, fmt.Errorf("unsupported query type")
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("no tables specified")
	}

	// For now, only handle single table queries
	if len(tables) > 1 {
		return nil, fmt.Errorf("JOIN queries not yet implemented")
	}

	// Read and parse CSV file
	tableName := tables[0]
	rows, err := readCSVFile(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", tableName, err)
	}

	if len(rows) == 0 {
		return NewResult([][]string{}), nil
	}

	// Get headers
	headers := rows[0]

	// Determine which columns to select
	selectedIndices := getSelectedColumnIndices(columns, headers)
	if selectedIndices == nil {
		return nil, fmt.Errorf("invalid column selection")
	}

	// Filter rows with WHERE clause and project columns
	result := [][]string{projectRow(headers, selectedIndices)}

	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Evaluate WHERE clause
		if whereExpr != nil {
			matches, err := eval.EvaluateExpression(whereExpr, row, headers)
			if err != nil {
				// Skip row on evaluation error
				continue
			}
			if !matches {
				continue
			}
		}

		// Project columns
		projectedRow := projectRow(row, selectedIndices)
		result = append(result, projectedRow)
	}

	return NewResult(result), nil
}

// readCSVFile reads a CSV file and returns all rows
func readCSVFile(filename string) ([][]string, error) {
	// Try to open the file as-is first, then try common paths
	paths := []string{
		filename,
		"tests/fixtures/" + filename,
		"./tests/fixtures/" + filename,
	}

	var file *os.File
	var err error

	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			// Found the file
			filename = path
			break
		}
	}

	if file == nil {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	defer file.Close()

	// Detect delimiter from file extension
	delimiter := ','
	if strings.HasSuffix(strings.ToLower(filename), ".tsv") ||
	   strings.HasSuffix(strings.ToLower(filename), ".tsv.gz") {
		delimiter = '\t'
	}

	// Check if file is gzipped
	var reader io.Reader = file
	if strings.HasSuffix(strings.ToLower(filename), ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Parse CSV data
	return parseCSVData(reader, rune(delimiter))
}

// getSelectedColumnIndices returns indices of selected columns
func getSelectedColumnIndices(columns []string, headers []string) []int {
	if columns == nil || len(columns) == 0 {
		return nil
	}

	// SELECT * - return all columns
	if len(columns) == 1 && columns[0] == "*" {
		indices := make([]int, len(headers))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// Specific columns
	var indices []int
	for _, col := range columns {
		col = strings.TrimSpace(col)
		found := false
		for i, header := range headers {
			if strings.EqualFold(header, col) {
				indices = append(indices, i)
				found = true
				break
			}
		}
		if !found {
			return nil // Column not found
		}
	}

	return indices
}

// projectRow selects specific columns from a row
func projectRow(row []string, indices []int) []string {
	result := make([]string, len(indices))
	for i, idx := range indices {
		if idx < len(row) {
			result[i] = row[idx]
		} else {
			result[i] = ""
		}
	}
	return result
}

// ExecuteSelectQueryWithData executes SELECT on provided CSV data
// This is used by contract tests
func ExecuteSelectQueryWithData(q query.Query, data io.Reader, delimiter rune) (Result, error) {
	if q == nil {
		return nil, fmt.Errorf("query is nil")
	}

	var columns []string
	var whereExpr query.Expression

	switch q.Type() {
	case query.QueryTypeSelect:
		sq := q.(*query.SelectQuery)
		columns = sq.Columns()
		whereExpr = sq.WhereClause()
	case query.QueryTypeJoin:
		jq := q.(*query.JoinQuery)
		columns = jq.Columns()
		whereExpr = jq.WhereClause()
	default:
		return nil, fmt.Errorf("unsupported query type: %v", q.Type())
	}

	// Parse CSV data
	rows, err := parseCSVData(data, delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(rows) == 0 {
		return NewResult([][]string{}), nil
	}

	// Get headers
	headers := rows[0]

	// Determine which columns to select
	selectedIndices := getSelectedColumnIndices(columns, headers)
	if selectedIndices == nil {
		return nil, fmt.Errorf("invalid column selection")
	}

	// Filter rows with WHERE clause and project columns
	result := [][]string{projectRow(headers, selectedIndices)}

	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Evaluate WHERE clause
		if whereExpr != nil {
			matches, err := eval.EvaluateExpression(whereExpr, row, headers)
			if err != nil {
				// Skip row on evaluation error
				continue
			}
			if !matches {
				continue
			}
		}

		// Project columns
		projectedRow := projectRow(row, selectedIndices)
		result = append(result, projectedRow)
	}

	return NewResult(result), nil
}

// parseCSVData parses CSV data from a reader
func parseCSVData(data io.Reader, delimiter rune) ([][]string, error) {
	// Read all data into buffer
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(data)
	if err != nil {
		return nil, err
	}

	// Create new reader from the buffered data for the parser
	content := buf.String()
	reader := bytes.NewReader([]byte(content))

	var rows [][]string
	parser := csv.NewCSVRowParser(reader, delimiter)

	for {
		row, err := parser.ParseRow()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(row) > 0 && !(len(row) == 1 && row[0] == "") {
			rows = append(rows, row)
		}
	}

	// If parser didn't work, fall back to simple line parsing
	if len(rows) == 0 {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Split(line, string(delimiter))
			rows = append(rows, fields)
		}
	}

	return rows, nil
}
