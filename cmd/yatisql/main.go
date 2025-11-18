package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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
