package eval

import (
	"fmt"

	"github.com/yourusername/yatisql/internal/query"
)

// EvaluateExpression evaluates a WHERE expression against a row of data
// Returns true if the expression matches, false otherwise
func EvaluateExpression(expr query.Expression, row []string, headers []string) (bool, error) {
	if expr == nil {
		return true, nil
	}

	// Handle binary expressions (AND, OR, comparisons)
	if binExpr, ok := expr.(*query.BinaryExpression); ok {
		return evaluateBinaryExpression(binExpr, row, headers)
	}

	return false, fmt.Errorf("unsupported expression type")
}

// evaluateBinaryExpression evaluates a binary expression
func evaluateBinaryExpression(expr *query.BinaryExpression, row []string, headers []string) (bool, error) {
	operator := expr.Operator()

	// Handle AND and OR operators
	if operator == "AND" {
		left, err := EvaluateExpression(expr.Left(), row, headers)
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil
		}
		right, err := EvaluateExpression(expr.Right(), row, headers)
		return right, err
	}

	if operator == "OR" {
		left, err := EvaluateExpression(expr.Left(), row, headers)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		right, err := EvaluateExpression(expr.Right(), row, headers)
		return right, err
	}

	// Handle comparison operators
	return evaluateComparison(expr, row, headers)
}

// evaluateComparison evaluates a comparison expression
func evaluateComparison(expr *query.BinaryExpression, row []string, headers []string) (bool, error) {
	// Get left value
	leftVal, err := resolveExpressionValue(expr.Left(), row, headers)
	if err != nil {
		return false, err
	}

	// Get right value
	rightVal, err := resolveExpressionValue(expr.Right(), row, headers)
	if err != nil {
		return false, err
	}

	// Compare values
	return compareValues(leftVal, rightVal, expr.Operator())
}

// resolveExpressionValue resolves an expression to a value
func resolveExpressionValue(expr query.Expression, row []string, headers []string) (string, error) {
	// Handle column references
	if colExpr, ok := expr.(*query.ColumnExpression); ok {
		colName := colExpr.Column()
		for i, header := range headers {
			if header == colName {
				if i < len(row) {
					return row[i], nil
				}
				return "", fmt.Errorf("column %s not found in row", colName)
			}
		}
		return "", fmt.Errorf("column %s not found in headers", colName)
	}

	// Handle literals
	if litExpr, ok := expr.(*query.LiteralExpression); ok {
		return fmt.Sprint(litExpr.Value()), nil
	}

	return "", fmt.Errorf("unsupported expression type for value resolution")
}

// compareValues compares two string values using an operator
func compareValues(left, right string, operator string) (bool, error) {
	lv := NewValue(left)
	rv := NewValue(right)

	switch operator {
	case "=":
		cmp, err := lv.CompareTo(rv)
		return cmp == 0, err

	case "!=", "<>":
		cmp, err := lv.CompareTo(rv)
		return cmp != 0, err

	case ">":
		cmp, err := lv.CompareTo(rv)
		return cmp > 0, err

	case "<":
		cmp, err := lv.CompareTo(rv)
		return cmp < 0, err

	case ">=":
		cmp, err := lv.CompareTo(rv)
		return cmp >= 0, err

	case "<=":
		cmp, err := lv.CompareTo(rv)
		return cmp <= 0, err

	case "LIKE":
		return likeCmp(left, right), nil

	case "IS NULL":
		return lv.IsNull, nil

	case "IS NOT NULL":
		return !lv.IsNull, nil

	default:
		return false, fmt.Errorf("unsupported comparison operator: %s", operator)
	}
}

// likeCmp performs LIKE comparison (simple pattern matching)
func likeCmp(value, pattern string) bool {
	// Very simplified LIKE: just check if pattern is contained in value
	// A full implementation would handle % and _ properly
	if pattern == "%" {
		return true
	}
	if pattern[0] != '%' && pattern[len(pattern)-1] != '%' {
		return value == pattern
	}
	return true
}
