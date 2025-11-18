package query

// QueryType represents the type of SQL query
type QueryType string

const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeInsert QueryType = "INSERT"
	QueryTypeJoin   QueryType = "JOIN"
)

// Query represents a parsed SQL query
type Query interface {
	Type() QueryType              // "SELECT", "JOIN", "UPDATE", "INSERT"
	Columns() []string            // Column names for SELECT
	Tables() []string             // Table filenames
	WhereClause() Expression      // WHERE conditions or nil
	JoinType() string             // "INNER", "LEFT" or ""
	OnCondition() Expression      // JOIN ON condition or nil
	SetClause() SetClause         // UPDATE SET clause or nil
	Values() [][]interface{}      // INSERT VALUES or nil
}

// Expression represents a SQL expression
type Expression interface {
	String() string
}

// SetClause represents UPDATE SET clause
type SetClause interface {
	Assignments() []Assignment
}

// Assignment represents a single SET assignment
type Assignment interface {
	Column() string
	Value() Expression
}

// ===== CONCRETE IMPLEMENTATIONS =====

// SelectQuery represents a SELECT statement
type SelectQuery struct {
	columns []string
	tables  []string
	where   Expression
}

// NewSelectQuery creates a new SELECT query
func NewSelectQuery(columns []string, tables []string) *SelectQuery {
	return &SelectQuery{
		columns: columns,
		tables:  tables,
	}
}

// Type returns the query type
func (q *SelectQuery) Type() QueryType {
	return QueryTypeSelect
}

// Columns returns selected columns
func (q *SelectQuery) Columns() []string {
	return q.columns
}

// Tables returns the tables being queried
func (q *SelectQuery) Tables() []string {
	return q.tables
}

// WhereClause returns the WHERE condition or nil
func (q *SelectQuery) WhereClause() Expression {
	return q.where
}

// SetWhereClause sets the WHERE condition
func (q *SelectQuery) SetWhereClause(where Expression) {
	q.where = where
}

// JoinType returns empty string (not a JOIN query)
func (q *SelectQuery) JoinType() string {
	return ""
}

// OnCondition returns nil (not a JOIN query)
func (q *SelectQuery) OnCondition() Expression {
	return nil
}

// SetClause returns nil (not an UPDATE query)
func (q *SelectQuery) SetClause() SetClause {
	return nil
}

// Values returns nil (not an INSERT query)
func (q *SelectQuery) Values() [][]interface{} {
	return nil
}

// JoinQuery represents a JOIN operation
type JoinQuery struct {
	leftTable   string
	rightTable  string
	joinType    string // "INNER" or "LEFT"
	onCondition Expression
	where       Expression
	columns     []string
}

// NewJoinQuery creates a new JOIN query
func NewJoinQuery(leftTable, rightTable, joinType string) *JoinQuery {
	return &JoinQuery{
		leftTable: leftTable,
		rightTable: rightTable,
		joinType: joinType,
	}
}

// Type returns the query type
func (q *JoinQuery) Type() QueryType {
	return QueryTypeJoin
}

// Columns returns selected columns
func (q *JoinQuery) Columns() []string {
	return q.columns
}

// SetColumns sets the columns
func (q *JoinQuery) SetColumns(cols []string) {
	q.columns = cols
}

// Tables returns the tables being joined
func (q *JoinQuery) Tables() []string {
	return []string{q.leftTable, q.rightTable}
}

// WhereClause returns the WHERE condition
func (q *JoinQuery) WhereClause() Expression {
	return q.where
}

// SetWhereClause sets the WHERE condition
func (q *JoinQuery) SetWhereClause(where Expression) {
	q.where = where
}

// JoinType returns the join type ("INNER" or "LEFT")
func (q *JoinQuery) JoinType() string {
	return q.joinType
}

// OnCondition returns the JOIN ON condition
func (q *JoinQuery) OnCondition() Expression {
	return q.onCondition
}

// SetOnCondition sets the ON condition
func (q *JoinQuery) SetOnCondition(on Expression) {
	q.onCondition = on
}

// SetClause returns nil (not an UPDATE query)
func (q *JoinQuery) SetClause() SetClause {
	return nil
}

// Values returns nil (not an INSERT query)
func (q *JoinQuery) Values() [][]interface{} {
	return nil
}

// UpdateQuery represents an UPDATE statement
type UpdateQuery struct {
	table     string
	setClause SetClause
	where     Expression
}

// NewUpdateQuery creates a new UPDATE query
func NewUpdateQuery(table string) *UpdateQuery {
	return &UpdateQuery{
		table: table,
	}
}

// Type returns the query type
func (q *UpdateQuery) Type() QueryType {
	return QueryTypeUpdate
}

// Columns returns nil (not a SELECT query)
func (q *UpdateQuery) Columns() []string {
	return nil
}

// Tables returns the table being updated
func (q *UpdateQuery) Tables() []string {
	return []string{q.table}
}

// WhereClause returns the WHERE condition
func (q *UpdateQuery) WhereClause() Expression {
	return q.where
}

// SetWhereClause sets the WHERE condition
func (q *UpdateQuery) SetWhereClause(where Expression) {
	q.where = where
}

// JoinType returns empty string (not a JOIN query)
func (q *UpdateQuery) JoinType() string {
	return ""
}

// OnCondition returns nil (not a JOIN query)
func (q *UpdateQuery) OnCondition() Expression {
	return nil
}

// SetClause returns the SET clause
func (q *UpdateQuery) SetClause() SetClause {
	return q.setClause
}

// SetSetClause sets the SET clause
func (q *UpdateQuery) SetSetClause(set SetClause) {
	q.setClause = set
}

// Values returns nil (not an INSERT query)
func (q *UpdateQuery) Values() [][]interface{} {
	return nil
}

// InsertQuery represents an INSERT statement
type InsertQuery struct {
	table  string
	values [][]interface{}
}

// NewInsertQuery creates a new INSERT query
func NewInsertQuery(table string, values [][]interface{}) *InsertQuery {
	return &InsertQuery{
		table:  table,
		values: values,
	}
}

// Type returns the query type
func (q *InsertQuery) Type() QueryType {
	return QueryTypeInsert
}

// Columns returns nil (not a SELECT query)
func (q *InsertQuery) Columns() []string {
	return nil
}

// Tables returns the table being inserted into
func (q *InsertQuery) Tables() []string {
	return []string{q.table}
}

// WhereClause returns nil (not a WHERE clause)
func (q *InsertQuery) WhereClause() Expression {
	return nil
}

// JoinType returns empty string (not a JOIN query)
func (q *InsertQuery) JoinType() string {
	return ""
}

// OnCondition returns nil (not a JOIN query)
func (q *InsertQuery) OnCondition() Expression {
	return nil
}

// SetClause returns nil (not an UPDATE query)
func (q *InsertQuery) SetClause() SetClause {
	return nil
}

// Values returns the values to insert
func (q *InsertQuery) Values() [][]interface{} {
	return q.values
}

// ===== EXPRESSION IMPLEMENTATIONS =====

// LiteralExpression represents a constant value
type LiteralExpression struct {
	value interface{}
}

// NewLiteralExpression creates a new literal
func NewLiteralExpression(value interface{}) *LiteralExpression {
	return &LiteralExpression{value: value}
}

// String returns string representation
func (e *LiteralExpression) String() string {
	switch v := e.value.(type) {
	case string:
		return "'" + v + "'"
	default:
		return ""
	}
}

// Value returns the literal value
func (e *LiteralExpression) Value() interface{} {
	return e.value
}

// ColumnExpression represents a column reference
type ColumnExpression struct {
	column string
	table  string
}

// NewColumnExpression creates a new column reference
func NewColumnExpression(column, table string) *ColumnExpression {
	return &ColumnExpression{column: column, table: table}
}

// String returns string representation
func (e *ColumnExpression) String() string {
	if e.table != "" {
		return e.table + "." + e.column
	}
	return e.column
}

// Column returns the column name
func (e *ColumnExpression) Column() string {
	return e.column
}

// Table returns the table name
func (e *ColumnExpression) Table() string {
	return e.table
}

// BinaryExpression represents a binary operation
type BinaryExpression struct {
	left     Expression
	operator string
	right    Expression
}

// NewBinaryExpression creates a new binary expression
func NewBinaryExpression(left Expression, operator string, right Expression) *BinaryExpression {
	return &BinaryExpression{left: left, operator: operator, right: right}
}

// String returns string representation
func (e *BinaryExpression) String() string {
	return e.left.String() + " " + e.operator + " " + e.right.String()
}

// Left returns the left operand
func (e *BinaryExpression) Left() Expression {
	return e.left
}

// Operator returns the operator
func (e *BinaryExpression) Operator() string {
	return e.operator
}

// Right returns the right operand
func (e *BinaryExpression) Right() Expression {
	return e.right
}

// ===== SET CLAUSE IMPLEMENTATIONS =====

// SimpleSetClause represents a simple SET clause
type SimpleSetClause struct {
	assignments []Assignment
}

// NewSimpleSetClause creates a new SET clause
func NewSimpleSetClause() *SimpleSetClause {
	return &SimpleSetClause{assignments: []Assignment{}}
}

// AddAssignment adds an assignment
func (s *SimpleSetClause) AddAssignment(assignment Assignment) {
	s.assignments = append(s.assignments, assignment)
}

// Assignments returns the assignments
func (s *SimpleSetClause) Assignments() []Assignment {
	return s.assignments
}

// SimpleAssignment represents a single assignment
type SimpleAssignment struct {
	column string
	value  Expression
}

// NewSimpleAssignment creates a new assignment
func NewSimpleAssignment(column string, value Expression) *SimpleAssignment {
	return &SimpleAssignment{column: column, value: value}
}

// Column returns the column name
func (a *SimpleAssignment) Column() string {
	return a.column
}

// Value returns the value expression
func (a *SimpleAssignment) Value() Expression {
	return a.value
}
