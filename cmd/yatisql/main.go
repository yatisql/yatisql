package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/yatisql/internal/executor"
	"github.com/yourusername/yatisql/internal/query"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: yatisql \"SELECT * FROM file.csv\"\n")
		os.Exit(1)
	}

	sql := args[0]

	// Preprocess SQL to convert file paths to table names
	// E.g., "SELECT * FROM tests/fixtures/sample.csv" -> "SELECT * FROM sample.csv"
	sql = preprocessSQL(sql)

	// Parse the query
	parser := query.NewParser()
	q, err := parser.Parse(sql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing query: %v\n", err)
		os.Exit(1)
	}

	// Execute the query
	exe := executor.NewQueryExecutor()
	result, err := exe.Execute(context.Background(), q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
		os.Exit(1)
	}

	// Format and output results
	if result == nil {
		fmt.Fprintf(os.Stderr, "Error: no result returned\n")
		os.Exit(1)
	}

	csv, err := result.ToCSV()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting results: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(csv)
}

// preprocessSQL converts file paths in FROM clauses to just the filename
// E.g., "SELECT * FROM tests/fixtures/sample.csv" -> "SELECT * FROM sample.csv"
func preprocessSQL(sql string) string {
	// Simple regex-free approach: look for FROM and extract filename
	upper := strings.ToUpper(sql)
	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		return sql
	}

	// Find the table name after FROM
	afterFrom := sql[fromIdx+4:]

	// Skip whitespace
	afterFrom = strings.TrimLeft(afterFrom, " \t")

	// Find the end of the table name (space, WHERE, semicolon, end of string)
	endIdx := len(afterFrom)
	for i, ch := range afterFrom {
		if ch == ' ' || ch == ';' {
			endIdx = i
			break
		}
	}

	tablePath := strings.TrimSpace(afterFrom[:endIdx])
	if tablePath == "" {
		return sql
	}

	// Extract just the filename from the path
	filename := filepath.Base(tablePath)

	// Replace the path with just the filename
	return sql[:fromIdx+4] + " " + filename + afterFrom[endIdx:]
}
