package query

type Search struct {
	Filters []FilterClause
	Sort    *SortClause
	Paging  *PagingClause
}

type Lookup struct {
	ID string
}
