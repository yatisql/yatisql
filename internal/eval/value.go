package eval

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yourusername/yatisql/internal/schema"
)

// Value represents a typed value in the query engine
type Value struct {
	Raw    string           // Original string representation
	Type   schema.ColumnType // Inferred type
	IsNull bool             // Whether value is NULL
}

// NewValue creates a new value
func NewValue(raw string) *Value {
	if raw == "" || raw == "NULL" || strings.ToUpper(raw) == "NULL" {
		return &Value{Raw: raw, Type: schema.TypeNull, IsNull: true}
	}

	inferredType := inferType(raw)
	return &Value{
		Raw:    raw,
		Type:   inferredType,
		IsNull: false,
	}
}

// inferType infers the type of a value from its string representation
func inferType(value string) schema.ColumnType {
	if value == "" || strings.ToUpper(value) == "NULL" {
		return schema.TypeNull
	}

	// Check boolean
	if value == "true" || value == "false" {
		return schema.TypeBoolean
	}

	// Check integer
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return schema.TypeInteger
	}

	// Check float
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return schema.TypeFloat
	}

	return schema.TypeString
}

// AsString returns the value as a string
func (v *Value) AsString() (string, error) {
	if v.IsNull {
		return "", nil
	}
	return v.Raw, nil
}

// AsInt returns the value as an integer
func (v *Value) AsInt() (int64, error) {
	if v.IsNull {
		return 0, fmt.Errorf("cannot convert NULL to integer")
	}

	i, err := strconv.ParseInt(v.Raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q to integer: %w", v.Raw, err)
	}
	return i, nil
}

// AsFloat returns the value as a float
func (v *Value) AsFloat() (float64, error) {
	if v.IsNull {
		return 0, fmt.Errorf("cannot convert NULL to float")
	}

	f, err := strconv.ParseFloat(v.Raw, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert %q to float: %w", v.Raw, err)
	}
	return f, nil
}

// AsBool returns the value as a boolean
func (v *Value) AsBool() (bool, error) {
	if v.IsNull {
		return false, fmt.Errorf("cannot convert NULL to boolean")
	}

	switch v.Raw {
	case "true", "TRUE", "1":
		return true, nil
	case "false", "FALSE", "0":
		return false, nil
	default:
		return false, fmt.Errorf("cannot convert %q to boolean", v.Raw)
	}
}

// CompareTo compares this value to another value
// Returns -1 if v < other, 0 if v == other, 1 if v > other
// Returns error if comparison is not possible
func (v *Value) CompareTo(other *Value) (int, error) {
	// NULL comparisons
	if v.IsNull || other.IsNull {
		if v.IsNull && other.IsNull {
			return 0, nil
		}
		return 0, fmt.Errorf("cannot compare NULL value")
	}

	// Try numeric comparison first
	if v.Type == schema.TypeInteger || v.Type == schema.TypeFloat {
		if other.Type == schema.TypeInteger || other.Type == schema.TypeFloat {
			vf, _ := strconv.ParseFloat(v.Raw, 64)
			of, _ := strconv.ParseFloat(other.Raw, 64)

			if vf < of {
				return -1, nil
			} else if vf > of {
				return 1, nil
			}
			return 0, nil
		}
	}

	// Fall back to string comparison
	if v.Raw < other.Raw {
		return -1, nil
	} else if v.Raw > other.Raw {
		return 1, nil
	}
	return 0, nil
}
