package eval

import (
	"fmt"
	"strconv"
	"strings"
)

// Comparison operators
const (
	OpEqual         = "="
	OpNotEqual      = "!="
	OpLess          = "<"
	OpLessEqual     = "<="
	OpGreater       = ">"
	OpGreaterEqual  = ">="
	OpLike          = "LIKE"
	OpIsNull        = "IS NULL"
	OpIsNotNull     = "IS NOT NULL"
)

// Compare compares two values using the specified operator
// Returns true if the comparison is true
func Compare(left, right, operator string) (bool, error) {
	return compareValues(left, right, operator)
}

// CompareNumeric compares two numeric values
func CompareNumeric(left, right, operator string) (bool, error) {
	leftF, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return false, fmt.Errorf("left operand %q is not numeric", left)
	}

	rightF, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return false, fmt.Errorf("right operand %q is not numeric", right)
	}

	switch operator {
	case OpEqual:
		return leftF == rightF, nil
	case OpNotEqual:
		return leftF != rightF, nil
	case OpLess:
		return leftF < rightF, nil
	case OpLessEqual:
		return leftF <= rightF, nil
	case OpGreater:
		return leftF > rightF, nil
	case OpGreaterEqual:
		return leftF >= rightF, nil
	default:
		return false, fmt.Errorf("unsupported operator for numeric comparison: %s", operator)
	}
}

// CompareString compares two string values
func CompareString(left, right, operator string) (bool, error) {
	switch operator {
	case OpEqual:
		return left == right, nil
	case OpNotEqual:
		return left != right, nil
	case OpLess:
		return left < right, nil
	case OpLessEqual:
		return left <= right, nil
	case OpGreater:
		return left > right, nil
	case OpGreaterEqual:
		return left >= right, nil
	case OpLike:
		return Like(left, right), nil
	default:
		return false, fmt.Errorf("unsupported operator for string comparison: %s", operator)
	}
}

// Like performs SQL LIKE pattern matching
// Simplified implementation supporting:
// - '%' matches any sequence of characters
// - '_' matches any single character
func Like(value, pattern string) bool {
	// Handle empty pattern
	if pattern == "" {
		return value == ""
	}

	// Handle % at start
	if pattern[0] == '%' {
		if len(pattern) == 1 {
			return true // "%" matches everything
		}
		// Recursive: try matching rest of pattern from each position
		for i := 0; i <= len(value); i++ {
			if Like(value[i:], pattern[1:]) {
				return true
			}
		}
		return false
	}

	// Handle _ (match single char)
	if pattern[0] == '_' {
		if len(value) == 0 {
			return false
		}
		return Like(value[1:], pattern[1:])
	}

	// Handle regular character
	if len(value) == 0 {
		return false
	}

	// Must match exactly if not a wildcard
	if value[0] != pattern[0] {
		return false
	}

	return Like(value[1:], pattern[1:])
}

// EqualNullSafe compares two values with three-valued logic (handling NULLs)
// Returns:
//   true if values are equal (or both NULL)
//   false if values are not equal
//   error if comparison not possible
func EqualNullSafe(left, right string) bool {
	if left == "" || strings.ToUpper(left) == "NULL" {
		if right == "" || strings.ToUpper(right) == "NULL" {
			return true // Both NULL = equal
		}
		return false // NULL != non-NULL
	}

	if right == "" || strings.ToUpper(right) == "NULL" {
		return false // non-NULL != NULL
	}

	return left == right
}
