package query

type Operator int

const (
	OperatorEq Operator = iota
	OperatorNeq
	OperatorGt
	OperatorGte
	OperatorLt
	OperatorLte
	OperatorIn
	OperatorContains
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

type FilterClause struct {
	Path     string
	Operator Operator
	Value    any
}

type SortField struct {
	Path  string
	Order SortOrder
}

type SortClause struct {
	Fields []SortField
}

type PagingClause struct {
	Offset int
	Limit  int
}
