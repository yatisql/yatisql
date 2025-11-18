# yatisql - SQL Query Engine for CSV/TSV Files

Yet another tabular inefficient SQL - A high-performance CLI tool for querying CSV and TSV files using standard SQL syntax without converting them to binary formats.

## Features

### Core Capabilities (MVP)
- **SELECT Queries**: Query CSV/TSV files with column selection and WHERE filtering
- **CSV/TSV Support**: Automatic format detection and parsing according to RFC 4180
- **Compression**: Transparent support for gzip-compressed files (.csv.gz, .tsv.gz)
- **Large File Support**: Streaming architecture handles files up to 1GB+ without memory limits
- **Type Inference**: Automatic detection of column types (String, Integer, Float, Boolean, NULL)

### Extended Capabilities
- **JOIN Operations**: INNER and LEFT JOINs between multiple CSV files with field equality conditions
- **Data Modification**: UPDATE and INSERT statements with atomic file persistence
- **Column Aliasing**: Disambiguate columns with table aliases (FROM a.csv u JOIN b.csv o)
- **Automatic Indexing**: SHA256-based indexes for performance optimization (P3)
- **Concurrent Queries**: Safe concurrent execution with RWMutex protection

## Quick Start

### Installation

```bash
git clone https://github.com/yourusername/yatisql.git
cd yatisql
make build
```

Binary will be created in `bin/yatisql`

### Basic Usage

```bash
# SELECT all columns from CSV file
./bin/yatisql "SELECT * FROM data.csv"

# SELECT specific columns with WHERE filter
./bin/yatisql "SELECT id, name FROM data.csv WHERE age > 30"

# SELECT from TSV file (auto-detected)
./bin/yatisql "SELECT * FROM data.tsv"

# SELECT from compressed file
./bin/yatisql "SELECT * FROM data.csv.gz"

# JOIN two CSV files
./bin/yatisql "SELECT u.name, o.amount FROM users.csv u JOIN orders.csv o ON u.id = o.user_id"

# UPDATE existing data
./bin/yatisql "UPDATE data.csv SET age = 35 WHERE id = 1"

# INSERT new rows
./bin/yatisql "INSERT INTO data.csv VALUES (5, 'John', 28, 'NYC')"
```

### Input Formats

**CSV File** (data.csv):
```
id,name,age,city
1,Alice,28,NYC
2,Bob,35,LA
3,Charlie,22,Chicago
```

**TSV File** (data.tsv):
```
id	name	age	city
1	Alice	28	NYC
2	Bob	35	LA
```

**Compressed File** (data.csv.gz):
- Automatically decompressed and queried
- Supported formats: gzip (.gz), bzip2 (.bz2)

## Architecture

### Library-First Design

The project follows a library-first architecture separating the SQL query engine from the CLI wrapper:

```
yatisql/
├── cmd/yatisql/          # CLI entry point (thin wrapper)
└── internal/             # Core query engine library
    ├── query/            # SQL parsing (QueryParser interface)
    ├── executor/         # Query execution (QueryExecutor interface)
    ├── csv/              # CSV file I/O with streaming
    ├── index/            # Index management (IndexManager interface)
    ├── schema/           # Type system and schema inference
    └── eval/             # Expression evaluation
```

### Core Components

#### 1. QueryParser (internal/query/)
- Parses SQL queries into Abstract Syntax Tree (AST)
- Uses xwb1989/sqlparser for SQL parsing
- Supports: SELECT, WHERE, JOIN, UPDATE, INSERT
- Contract-tested API for unit testing

#### 2. QueryExecutor (internal/executor/)
- Executes parsed queries on CSV files
- Implements streaming algorithm for memory efficiency
- Routes to SELECT, JOIN, UPDATE, INSERT handlers
- Non-blocking concurrent query execution

#### 3. CSV Handler (internal/csv/)
- Streaming row reader (RowIterator interface)
- RFC 4180 CSV parser with proper escaping
- CSV row writer preserving original format
- Transparent compression support
- File-level mutation safety

#### 4. Index Manager (internal/index/)
- SHA256-based index validation
- Automatic index creation and caching
- Non-blocking index rebuild on file changes
- RWMutex for concurrent access
- Binary index file format with HashMap

#### 5. Type System (internal/schema/ + internal/eval/)
- Automatic type inference
- Type coercion with NULL handling
- Expression evaluation (WHERE, ON conditions)
- Comparison operations across types

## Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| SELECT (100MB) | < 5 sec | Streaming read |
| JOIN (50MB+50MB) | < 10 sec | First run without index |
| JOIN (indexed) | < 1 sec | With index acceleration |
| UPDATE/INSERT | < 2 seconds | Atomic file write |
| Large files | Unlimited | Streaming, no memory limits |

## Development

### Project Structure

```
yatisql/
├── cmd/
│   └── yatisql/main.go           # CLI entry point
├── internal/
│   ├── query/                    # SQL parsing
│   │   ├── parser.go
│   │   ├── lexer.go
│   │   └── ast.go
│   ├── executor/                 # Query execution
│   │   ├── executor.go
│   │   ├── select.go
│   │   ├── join.go
│   │   ├── update.go
│   │   └── insert.go
│   ├── csv/                      # File I/O
│   │   ├── file.go
│   │   ├── reader.go
│   │   ├── parser.go
│   │   ├── writer.go
│   │   └── compression.go
│   ├── index/                    # Indexing
│   │   ├── index.go
│   │   ├── manager.go
│   │   ├── builder.go
│   │   └── store.go
│   ├── schema/                   # Type system
│   │   └── column.go
│   └── eval/                     # Expression evaluation
│       ├── value.go
│       ├── condition.go
│       ├── compare.go
│       └── types.go
├── tests/
│   ├── contract/                 # API contract tests
│   ├── integration/              # End-to-end tests
│   ├── unit/                     # Unit tests
│   └── fixtures/                 # Test CSV/TSV files
├── specs/                        # Design documents
├── Makefile
├── go.mod
└── README.md
```

### Test Organization (TDD)

Tests follow Test-Driven Development methodology:

1. **Contract Tests** (`tests/contract/`)
   - QueryParser contracts
   - QueryExecutor contracts
   - IndexManager contracts
   - Written before implementation, must fail (RED)

2. **Integration Tests** (`tests/integration/`)
   - End-to-end user story testing
   - SELECT, JOIN, UPDATE, INSERT workflows
   - Large file and compression scenarios

3. **Unit Tests** (`tests/unit/`)
   - CSV parsing edge cases
   - Type coercion
   - Index validation

### Building and Testing

```bash
# Build the binary
make build

# Run all tests
make test

# Run tests with coverage
make test-cover

# Run linters
make lint

# Run benchmarks
make bench

# Clean build artifacts
make clean
```

### API Contracts

The library exposes three main interfaces:

#### QueryParser
```go
type QueryParser interface {
    Parse(queryStr string) (Query, error)
}
```

#### QueryExecutor
```go
type QueryExecutor interface {
    Execute(ctx context.Context, query Query, tables ...*CSVFile) (Result, error)
}
```

#### IndexManager
```go
type IndexManager interface {
    GetOrCreateIndex(ctx context.Context, file *CSVFile, column string) (*Index, error)
    IsValid(index *Index, file *CSVFile) bool
    RebuildIndex(ctx context.Context, file *CSVFile, column string) error
    DeleteIndex(file *CSVFile, column string) error
}
```

## Design Decisions

### Technology Stack
- **Language**: Go 1.21+ for performance, concurrency, and simplicity
- **SQL Parser**: xwb1989/sqlparser (lightweight, MySQL-compatible)
- **CSV Parsing**: Go stdlib `encoding/csv` (RFC 4180 compliant)
- **Compression**: Go stdlib `compress/gzip`, `compress/bzip2`
- **Concurrency**: Go goroutines, channels, RWMutex
- **Hashing**: Go stdlib `crypto/sha256` for index validation

### Key Principles
- **Library-First**: Query engine is standalone library, testable without CLI
- **Test-Driven Development**: All APIs tested before implementation
- **Contract-Driven**: Clear, documented interfaces
- **Streaming Architecture**: Memory bounded by result set size, not file size
- **Simplicity**: MVP focuses on core SQL operations; advanced features deferred

## Implementation Phases

### Phase 1: Setup (6 tasks)
- Go module, directory structure, Makefile, .gitignore, README, test fixtures

### Phase 2: Foundation (19 tasks, BLOCKING)
- CSV parsing, SQL parsing, type system, indexing infrastructure
- All components independently testable

### Phase 3: User Story 1 (8 tasks) - MVP
- SELECT queries with column projection and WHERE filtering
- TSV, compressed file support
- Deliverable: v0.1.0 with SELECT functionality

### Phase 4: User Story 2 (6 tasks)
- JOIN operations (INNER, LEFT)
- Column aliasing support

### Phase 5: User Story 3 (6 tasks)
- UPDATE and INSERT operations
- Atomic file persistence

### Phase 6: User Story 4 (4 tasks)
- Automatic index management
- Performance optimization

### Phase 7: Polish (13 tasks)
- Edge case tests, large file handling, concurrency tests
- Performance benchmarking, documentation
- Release v1.0.0 with full feature set

## Constitution Alignment

This project follows the **yatisql Constitution v1.0.0**:

- ✅ **I. Library-First**: Query engine is modular library
- ✅ **II. Test-Driven**: All tests written before implementation
- ✅ **III. Contract-Driven**: Public interfaces fully specified
- ✅ **IV. Semantic Versioning**: MAJOR.MINOR.PATCH versioning
- ✅ **V. Simplicity**: MVP focuses on core operations

## Contributing

Contributions welcome! Please follow the development guidelines:

1. Follow TDD: Write contract tests before implementation
2. Use library-first architecture for all components
3. Ensure 80%+ test coverage for new code
4. Document public APIs
5. Update tasks.md when adding/removing features

## License

Licensed under MIT License (see LICENSE file)

## References

- **Feature Spec**: `specs/001-sql-csv-query/spec.md`
- **Implementation Plan**: `specs/001-sql-csv-query/plan.md`
- **Data Model**: `specs/001-sql-csv-query/data-model.md`
- **API Contracts**: `specs/001-sql-csv-query/contracts/`
- **Research**: `specs/001-sql-csv-query/research.md`
- **Task List**: `specs/001-sql-csv-query/tasks.md`

## Contact

For questions or issues, please open a GitHub issue or contact the maintainers.

---

**Status**: MVP Phase (Phase 1-3) - SELECT functionality implemented
**Next**: Phase 4 - JOIN operations
