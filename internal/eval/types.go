package eval

import (
	"fmt"
	"strconv"

	"github.com/yourusername/yatisql/internal/schema"
)

// InferType determines the type of a value from its string representation
func InferType(value string) schema.ColumnType {
	return inferType(value)
}

// CoerceValue converts a value to a target type
func CoerceValue(value string, targetType schema.ColumnType) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	switch targetType {
	case schema.TypeString:
		return value, nil

	case schema.TypeInteger:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to INTEGER", value)
		}
		return int(i), nil

	case schema.TypeFloat:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to FLOAT", value)
		}
		return f, nil

	case schema.TypeBoolean:
		switch value {
		case "true", "TRUE", "1":
			return true, nil
		case "false", "FALSE", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("cannot convert %q to BOOLEAN", value)
		}

	case schema.TypeNull:
		return nil, nil

	default:
		return value, nil
	}
}

// PromoteType returns the wider type that can represent both types
func PromoteType(t1, t2 schema.ColumnType) schema.ColumnType {
	if t1 == t2 {
		return t1
	}

	// NULL promotes to the other type
	if t1 == schema.TypeNull {
		return t2
	}
	if t2 == schema.TypeNull {
		return t1
	}

	// FLOAT can represent INTEGER
	if (t1 == schema.TypeFloat && t2 == schema.TypeInteger) ||
		(t1 == schema.TypeInteger && t2 == schema.TypeFloat) {
		return schema.TypeFloat
	}

	// Everything else promotes to STRING
	return schema.TypeString
}

// CanCompare checks if two types can be compared
func CanCompare(t1, t2 schema.ColumnType) bool {
	// Numeric types can be compared
	if isNumeric(t1) && isNumeric(t2) {
		return true
	}

	// String with anything
	if t1 == schema.TypeString || t2 == schema.TypeString {
		return true
	}

	// Same types
	if t1 == t2 {
		return true
	}

	return false
}

// isNumeric checks if a type is numeric
func isNumeric(t schema.ColumnType) bool {
	return t == schema.TypeInteger || t == schema.TypeFloat
}
