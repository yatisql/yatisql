package contract

import (
	"testing"

	"github.com/yourusername/yatisql/internal/query"
)

// TestParserSELECTSimple tests parsing simple SELECT queries
func TestParserSELECTSimple(t *testing.T) {
	sql := "SELECT id, name FROM users.csv"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	// Verify it's a SELECT query
	if string(q.Type()) != "SELECT" {
		t.Fatalf("Expected SELECT query, got %s", q.Type())
	}

	// Verify columns
	columns := q.Columns()
	if len(columns) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(columns))
	}
	if columns[0] != "id" || columns[1] != "name" {
		t.Errorf("Expected [id, name], got %v", columns)
	}

	// Verify table
	tables := q.Tables()
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}
	if tables[0] != "users.csv" {
		t.Errorf("Expected users.csv, got %s", tables[0])
	}
}

// TestParserSELECTAll tests SELECT * queries
func TestParserSELECTAll(t *testing.T) {
	sql := "SELECT * FROM data.csv"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "SELECT" {
		t.Fatalf("Expected SELECT query")
	}

	// Verify all columns
	columns := q.Columns()
	if len(columns) != 1 || columns[0] != "*" {
		t.Errorf("Expected [*], got %v", columns)
	}
}

// TestParserSELECTWHERE tests SELECT with WHERE clause
func TestParserSELECTWHERE(t *testing.T) {
	sql := "SELECT * FROM users.csv WHERE age > 30"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "SELECT" {
		t.Fatalf("Expected SELECT query")
	}

	// Verify WHERE clause exists
	whereClause := q.WhereClause()
	if whereClause == nil {
		t.Fatalf("Expected WHERE clause")
	}
	// The parser converts numeric literals to quoted strings
	if whereClause.String() != "age > '30'" && whereClause.String() != "age > 30" {
		t.Errorf("Expected 'age > 30' or 'age > '30'', got %s", whereClause.String())
	}
}

// TestParserSELECTMultipleConditions tests SELECT with complex WHERE
func TestParserSELECTMultipleConditions(t *testing.T) {
	sql := "SELECT * FROM users.csv WHERE age > 30 AND city = 'NYC'"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	whereClause := q.WhereClause()
	if whereClause == nil {
		t.Fatalf("Expected WHERE clause")
	}
	// Should contain both conditions
	if whereClause.String() == "" {
		t.Errorf("WHERE clause is empty")
	}
}

// TestParserINNERJOIN tests INNER JOIN parsing
func TestParserINNERJOIN(t *testing.T) {
	sql := "SELECT u.id, u.name, o.amount FROM users.csv u INNER JOIN orders.csv o ON u.id = o.user_id"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "JOIN" {
		t.Fatalf("Expected JOIN query, got %s", q.Type())
	}

	// Verify tables
	tables := q.Tables()
	if len(tables) != 2 {
		t.Fatalf("Expected 2 tables, got %d", len(tables))
	}

	// Verify JOIN type
	joinType := q.JoinType()
	if joinType != "INNER" {
		t.Errorf("Expected INNER join, got %s", joinType)
	}

	// Verify ON condition
	onCondition := q.OnCondition()
	if onCondition == nil {
		t.Fatalf("Expected ON condition")
	}
	if onCondition.String() != "u.id = o.user_id" {
		t.Errorf("Expected 'u.id = o.user_id', got %s", onCondition.String())
	}
}

// TestParserLEFTJOIN tests LEFT JOIN parsing
func TestParserLEFTJOIN(t *testing.T) {
	sql := "SELECT * FROM users.csv u LEFT JOIN orders.csv o ON u.id = o.user_id"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "JOIN" {
		t.Fatalf("Expected JOIN query")
	}

	joinType := q.JoinType()
	if joinType != "LEFT" {
		t.Errorf("Expected LEFT join, got %s", joinType)
	}
}

// TestParserJOINWithWHERE tests JOIN with WHERE filtering
func TestParserJOINWithWHERE(t *testing.T) {
	sql := "SELECT * FROM users.csv u INNER JOIN orders.csv o ON u.id = o.user_id WHERE o.amount > 100"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "JOIN" {
		t.Fatalf("Expected JOIN query")
	}

	whereClause := q.WhereClause()
	if whereClause == nil {
		t.Fatalf("Expected WHERE clause")
	}
	// The parser converts numeric literals to quoted strings
	if whereClause.String() != "o.amount > '100'" && whereClause.String() != "o.amount > 100" {
		t.Errorf("Expected 'o.amount > 100' or 'o.amount > '100'', got %s", whereClause.String())
	}
}

// TestParserUPDATE tests UPDATE statement parsing
func TestParserUPDATE(t *testing.T) {
	sql := "UPDATE users.csv SET age = 35 WHERE id = 1"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "UPDATE" {
		t.Fatalf("Expected UPDATE query, got %s", q.Type())
	}

	// Verify table
	tables := q.Tables()
	if len(tables) != 1 || tables[0] != "users.csv" {
		t.Errorf("Expected [users.csv], got %v", tables)
	}

	// Verify SET clause
	setClause := q.SetClause()
	if setClause == nil {
		t.Fatalf("Expected SET clause")
	}
	// Should have age = 35
	assignments := setClause.Assignments()
	if len(assignments) != 1 {
		t.Fatalf("Expected 1 assignment, got %d", len(assignments))
	}
	if assignments[0].Column() != "age" {
		t.Errorf("Expected column 'age', got %s", assignments[0].Column())
	}
}

// TestParserUPDATEMultiple tests UPDATE with multiple SET clauses
func TestParserUPDATEMultiple(t *testing.T) {
	sql := "UPDATE users.csv SET age = 30, city = 'LA' WHERE id = 1"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "UPDATE" {
		t.Fatalf("Expected UPDATE query")
	}

	setClause := q.SetClause()
	assignments := setClause.Assignments()
	if len(assignments) != 2 {
		t.Fatalf("Expected 2 assignments, got %d", len(assignments))
	}
}

// TestParserINSERT tests INSERT statement parsing
func TestParserINSERT(t *testing.T) {
	sql := "INSERT INTO users.csv VALUES (5, 'John', 28, 'NYC')"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "INSERT" {
		t.Fatalf("Expected INSERT query, got %s", q.Type())
	}

	// Verify table
	tables := q.Tables()
	if len(tables) != 1 || tables[0] != "users.csv" {
		t.Errorf("Expected [users.csv], got %v", tables)
	}

	// Verify VALUES clause
	values := q.Values()
	if len(values) == 0 {
		t.Fatalf("Expected VALUES")
	}
}

// TestParserINSERTMultipleRows tests INSERT with multiple rows
func TestParserINSERTMultipleRows(t *testing.T) {
	sql := "INSERT INTO users.csv VALUES (5, 'John', 28, 'NYC'), (6, 'Jane', 26, 'LA')"
	q, err := query.ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}

	if q.Type() != "INSERT" {
		t.Fatalf("Expected INSERT query")
	}

	values := q.Values()
	if len(values) != 2 {
		t.Fatalf("Expected 2 value sets, got %d", len(values))
	}
}

// TestParserEdgeCases tests edge cases in SQL parsing
func TestParserEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		sql      string
		queryType string
	}{
		{"SELECT with aliases", "SELECT id AS user_id FROM users.csv", "SELECT"},
		{"SELECT with quoted string", "SELECT * FROM users.csv WHERE name = 'O''Brien'", "SELECT"},
		{"Column with table alias", "SELECT u.name FROM users.csv u", "SELECT"},
		{"UPDATE without WHERE", "UPDATE users.csv SET age = 30", "UPDATE"},
		{"INSERT with different value counts", "INSERT INTO users.csv VALUES (1, 'John')", "INSERT"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := query.ParseSQL(tc.sql)
			if err != nil {
				t.Fatalf("ParseSQL failed: %v", err)
			}
			if string(q.Type()) != tc.queryType {
				t.Errorf("Expected %s, got %s", tc.queryType, q.Type())
			}
		})
	}
}

