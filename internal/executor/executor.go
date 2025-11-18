package executor

import (
	"context"
	"fmt"

	"github.com/yourusername/yatisql/internal/query"
)

// QueryExecutor executes parsed SQL queries
type QueryExecutor interface {
	Execute(ctx context.Context, q query.Query, options ...ExecuteOption) (Result, error)
}

// ExecuteOption is an option for execution
type ExecuteOption func(*executionConfig)

type executionConfig struct {
	useIndex bool
}

// Result represents query execution results
type Result interface {
	Rows() [][]string               // All rows including header
	ToCSV() (string, error)         // Format as CSV
	ToJSON() (string, error)        // Format as JSON
	ToTable() (string, error)       // Format as ASCII table
	ColumnCount() int               // Number of columns
	RowCount() int                  // Number of data rows (excluding header)
}

// simpleExecutor implements QueryExecutor
type simpleExecutor struct {
	// Could add config, loggers, etc.
}

// NewQueryExecutor creates a new query executor
func NewQueryExecutor() QueryExecutor {
	return &simpleExecutor{}
}

// Execute executes a query
func (e *simpleExecutor) Execute(ctx context.Context, q query.Query, options ...ExecuteOption) (Result, error) {
	if q == nil {
		return nil, fmt.Errorf("query is nil")
	}

	switch query.QueryType(q.Type()) {
	case query.QueryTypeSelect:
		return ExecuteSelectQuery(q)
	case query.QueryTypeUpdate:
		return nil, fmt.Errorf("UPDATE not yet implemented")
	case query.QueryTypeInsert:
		return nil, fmt.Errorf("INSERT not yet implemented")
	case query.QueryTypeJoin:
		return ExecuteSelectQuery(q) // JOIN is a variant of SELECT
	default:
		return nil, fmt.Errorf("unsupported query type: %s", q.Type())
	}
}

// SimpleResult implements Result interface
type SimpleResult struct {
	rows [][]string
}

// NewResult creates a new result
func NewResult(rows [][]string) *SimpleResult {
	return &SimpleResult{rows: rows}
}

// Rows returns all rows
func (r *SimpleResult) Rows() [][]string {
	return r.rows
}

// ColumnCount returns number of columns
func (r *SimpleResult) ColumnCount() int {
	if len(r.rows) == 0 {
		return 0
	}
	return len(r.rows[0])
}

// RowCount returns number of data rows (excluding header)
func (r *SimpleResult) RowCount() int {
	if len(r.rows) <= 1 {
		return 0
	}
	return len(r.rows) - 1
}

// ToCSV formats result as CSV
func (r *SimpleResult) ToCSV() (string, error) {
	if len(r.rows) == 0 {
		return "", nil
	}

	var output string
	for i, row := range r.rows {
		for j, field := range row {
			if j > 0 {
				output += ","
			}
			// Simple CSV escaping
			if needsQuotes(field) {
				output += `"` + escapeQuotes(field) + `"`
			} else {
				output += field
			}
		}
		if i < len(r.rows)-1 {
			output += "\n"
		}
	}

	return output, nil
}

// ToJSON formats result as JSON
func (r *SimpleResult) ToJSON() (string, error) {
	if len(r.rows) == 0 {
		return "[]", nil
	}

	output := "[\n"
	headers := r.rows[0]

	for i := 1; i < len(r.rows); i++ {
		row := r.rows[i]
		output += "  {"

		for j, header := range headers {
			if j > 0 {
				output += ", "
			}
			value := ""
			if j < len(row) {
				value = row[j]
			}
			output += fmt.Sprintf(`"%s": "%s"`, header, escapeJSON(value))
		}

		output += "}"
		if i < len(r.rows)-1 {
			output += ","
		}
		output += "\n"
	}

	output += "]"
	return output, nil
}

// ToTable formats result as ASCII table
func (r *SimpleResult) ToTable() (string, error) {
	if len(r.rows) == 0 {
		return "", nil
	}

	// Simple ASCII table formatting
	output := ""
	for _, row := range r.rows {
		for j, field := range row {
			if j > 0 {
				output += " | "
			}
			output += field
		}
		output += "\n"
	}

	return output, nil
}

// needsQuotes checks if a field needs CSV quoting
func needsQuotes(field string) bool {
	for _, ch := range field {
		if ch == ',' || ch == '"' || ch == '\n' {
			return true
		}
	}
	return false
}

// escapeQuotes escapes quotes in a field for CSV
func escapeQuotes(field string) string {
	var output string
	for _, ch := range field {
		if ch == '"' {
			output += `""`
		} else {
			output += string(ch)
		}
	}
	return output
}

// escapeJSON escapes special characters for JSON
func escapeJSON(s string) string {
	var output string
	for _, ch := range s {
		switch ch {
		case '"':
			output += `\"`
		case '\\':
			output += `\\`
		case '\n':
			output += `\n`
		case '\r':
			output += `\r`
		case '\t':
			output += `\t`
		default:
			output += string(ch)
		}
	}
	return output
}
