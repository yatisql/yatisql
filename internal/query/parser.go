package query

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

// ParseSQL parses a SQL query string and returns a Query object
// Supports:
// - SELECT with column projection, WHERE filtering, JOIN operations, aliases
// - UPDATE with SET and WHERE clauses
// - INSERT with VALUES
func ParseSQL(sql string) (Query, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("SQL parse error: %w", err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		return parseSelectQuery(stmt)
	case *sqlparser.Update:
		return parseUpdateQuery(stmt)
	case *sqlparser.Insert:
		return parseInsertQuery(stmt)
	default:
		return nil, fmt.Errorf("unsupported query type: %T", stmt)
	}
}

// parseSelectQuery converts sqlparser.Select to our SelectQuery
func parseSelectQuery(stmt *sqlparser.Select) (Query, error) {
	// Extract columns
	columns := []string{}
	for _, expr := range stmt.SelectExprs {
		switch expr := expr.(type) {
		case *sqlparser.StarExpr:
			columns = append(columns, "*")
		case *sqlparser.AliasedExpr:
			col := sqlparser.String(expr.Expr)
			columns = append(columns, col)
		}
	}

	// Extract tables
	tables := []string{}
	if stmt.From != nil {
		for _, tableExpr := range stmt.From {
			switch tableExpr := tableExpr.(type) {
			case *sqlparser.AliasedTableExpr:
				tableName := sqlparser.String(tableExpr.Expr)
				tables = append(tables, tableName)
			case *sqlparser.JoinTableExpr:
				// Handle JOIN
				return parseJoinQuery(stmt, tableExpr)
			}
		}
	}

	query := NewSelectQuery(columns, tables)

	// Parse WHERE clause if present
	if stmt.Where != nil {
		where := parseExpression(stmt.Where.Expr)
		query.SetWhereClause(where)
	}

	return query, nil
}

// parseJoinQuery converts a JOIN to our JoinQuery
func parseJoinQuery(stmt *sqlparser.Select, joinTable *sqlparser.JoinTableExpr) (Query, error) {
	// Get left table
	var leftTable string
	if aliasedExpr, ok := joinTable.LeftExpr.(*sqlparser.AliasedTableExpr); ok {
		leftTable = sqlparser.String(aliasedExpr.Expr)
	} else {
		leftTable = sqlparser.String(joinTable.LeftExpr)
	}

	// Get right table
	var rightTable string
	if aliasedExpr, ok := joinTable.RightExpr.(*sqlparser.AliasedTableExpr); ok {
		rightTable = sqlparser.String(aliasedExpr.Expr)
	} else {
		rightTable = sqlparser.String(joinTable.RightExpr)
	}

	// Get join type
	joinType := "INNER"
	if joinTable.Join != "" {
		if strings.Contains(strings.ToUpper(joinTable.Join), "LEFT") {
			joinType = "LEFT"
		}
	}

	query := NewJoinQuery(leftTable, rightTable, joinType)

	// Extract columns from SELECT
	columns := []string{}
	for _, expr := range stmt.SelectExprs {
		switch expr := expr.(type) {
		case *sqlparser.StarExpr:
			columns = append(columns, "*")
		case *sqlparser.AliasedExpr:
			col := sqlparser.String(expr.Expr)
			columns = append(columns, col)
		}
	}
	query.SetColumns(columns)

	// Parse ON condition (stored in Condition field)
	if joinTable.Condition.On != nil {
		onExpr := parseExpression(joinTable.Condition.On)
		query.SetOnCondition(onExpr)
	}

	// Parse WHERE clause if present
	if stmt.Where != nil {
		where := parseExpression(stmt.Where.Expr)
		query.SetWhereClause(where)
	}

	return query, nil
}

// parseUpdateQuery converts sqlparser.Update to our UpdateQuery
func parseUpdateQuery(stmt *sqlparser.Update) (Query, error) {
	var tableName string
	if len(stmt.TableExprs) > 0 {
		tableName = sqlparser.String(stmt.TableExprs[0])
	}

	query := NewUpdateQuery(tableName)

	// Parse SET clause
	setClause := NewSimpleSetClause()
	for _, update := range stmt.Exprs {
		colName := sqlparser.String(update.Name)
		value := parseExpression(update.Expr)
		assignment := NewSimpleAssignment(colName, value)
		setClause.AddAssignment(assignment)
	}
	query.SetSetClause(setClause)

	// Parse WHERE clause if present
	if stmt.Where != nil {
		where := parseExpression(stmt.Where.Expr)
		query.SetWhereClause(where)
	}

	return query, nil
}

// parseInsertQuery converts sqlparser.Insert to our InsertQuery
func parseInsertQuery(stmt *sqlparser.Insert) (Query, error) {
	tableName := sqlparser.String(stmt.Table)

	// Extract values
	values := [][]interface{}{}
	if stmt.Rows != nil {
		switch rows := stmt.Rows.(type) {
		case sqlparser.Values:
			for _, row := range rows {
				rowValues := []interface{}{}
				for _, val := range row {
					rowValues = append(rowValues, sqlparser.String(val))
				}
				values = append(values, rowValues)
			}
		}
	}

	return NewInsertQuery(tableName, values), nil
}

// parseExpression converts sqlparser expressions to our Expression interface
func parseExpression(expr sqlparser.Expr) Expression {
	switch expr := expr.(type) {
	case *sqlparser.ComparisonExpr:
		left := parseExpression(expr.Left)
		right := parseExpression(expr.Right)
		operator := expr.Operator
		return NewBinaryExpression(left, operator, right)

	case *sqlparser.AndExpr:
		left := parseExpression(expr.Left)
		right := parseExpression(expr.Right)
		return NewBinaryExpression(left, "AND", right)

	case *sqlparser.OrExpr:
		left := parseExpression(expr.Left)
		right := parseExpression(expr.Right)
		return NewBinaryExpression(left, "OR", right)

	case *sqlparser.ColName:
		table := sqlparser.String(expr.Qualifier)
		column := expr.Name.String()
		return NewColumnExpression(column, table)

	case *sqlparser.SQLVal:
		// Extract literal value
		val := expr.Val
		return NewLiteralExpression(string(val))

	default:
		// Default: convert to string
		return NewLiteralExpression(sqlparser.String(expr))
	}
}
