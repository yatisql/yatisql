package schema

import "fmt"

// ColumnType represents the data type of a column
type ColumnType int

const (
	TypeUnknown ColumnType = iota
	TypeString
	TypeInteger
	TypeFloat
	TypeBoolean
	TypeNull
)

// String returns string representation of type
func (ct ColumnType) String() string {
	switch ct {
	case TypeString:
		return "STRING"
	case TypeInteger:
		return "INTEGER"
	case TypeFloat:
		return "FLOAT"
	case TypeBoolean:
		return "BOOLEAN"
	case TypeNull:
		return "NULL"
	default:
		return "UNKNOWN"
	}
}

// Column represents a database column with metadata
type Column struct {
	Name     string
	Index    int
	Type     ColumnType
	Nullable bool
}

// NewColumn creates a new column
func NewColumn(name string, index int, colType ColumnType) *Column {
	return &Column{
		Name:     name,
		Index:    index,
		Type:     colType,
		Nullable: true,
	}
}

// InferType infers the column type from a string value
func (c *Column) InferType(value string) ColumnType {
	if value == "" || value == "NULL" {
		return TypeNull
	}

	// Try to parse as boolean
	if value == "true" || value == "false" {
		return TypeBoolean
	}

	// Try to parse as integer
	if _, err := fmt.Sscanf(value, "%d", new(int)); err == nil {
		return TypeInteger
	}

	// Try to parse as float
	if _, err := fmt.Sscanf(value, "%f", new(float64)); err == nil {
		return TypeFloat
	}

	// Default to string
	return TypeString
}

// Coerce converts a value to the column's type
func (c *Column) Coerce(value interface{}) (interface{}, error) {
	if value == nil || value == "" {
		if !c.Nullable {
			return nil, fmt.Errorf("column %s does not allow NULL", c.Name)
		}
		return nil, nil
	}

	strVal := fmt.Sprint(value)

	switch c.Type {
	case TypeString:
		return strVal, nil

	case TypeInteger:
		var i int
		if _, err := fmt.Sscanf(strVal, "%d", &i); err != nil {
			return nil, fmt.Errorf("cannot convert %q to INTEGER: %w", strVal, err)
		}
		return i, nil

	case TypeFloat:
		var f float64
		if _, err := fmt.Sscanf(strVal, "%f", &f); err != nil {
			return nil, fmt.Errorf("cannot convert %q to FLOAT: %w", strVal, err)
		}
		return f, nil

	case TypeBoolean:
		switch strVal {
		case "true", "TRUE", "1":
			return true, nil
		case "false", "FALSE", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("cannot convert %q to BOOLEAN", strVal)
		}

	default:
		return strVal, nil
	}
}
