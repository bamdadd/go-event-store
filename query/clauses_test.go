package query_test

import (
	"testing"

	"github.com/bamdadd/go-event-store/query"
	"github.com/stretchr/testify/assert"
)

func TestFilterClauseConstruction(t *testing.T) {
	f := query.FilterClause{
		Path:     "state.status",
		Operator: query.OperatorEq,
		Value:    "active",
	}
	assert.Equal(t, "state.status", f.Path)
	assert.Equal(t, query.OperatorEq, f.Operator)
	assert.Equal(t, "active", f.Value)
}

func TestSortClauseConstruction(t *testing.T) {
	s := query.SortClause{
		Fields: []query.SortField{
			{Path: "metadata.updated_at", Order: query.SortDesc},
		},
	}
	assert.Len(t, s.Fields, 1)
	assert.Equal(t, query.SortDesc, s.Fields[0].Order)
}

func TestPagingClauseConstruction(t *testing.T) {
	p := query.PagingClause{Offset: 10, Limit: 25}
	assert.Equal(t, 10, p.Offset)
	assert.Equal(t, 25, p.Limit)
}

func TestSearchConstruction(t *testing.T) {
	s := query.Search{
		Filters: []query.FilterClause{
			{Path: "state.total", Operator: query.OperatorGt, Value: 100},
		},
		Sort: &query.SortClause{
			Fields: []query.SortField{{Path: "state.total", Order: query.SortAsc}},
		},
		Paging: &query.PagingClause{Offset: 0, Limit: 50},
	}
	assert.Len(t, s.Filters, 1)
	assert.NotNil(t, s.Sort)
	assert.NotNil(t, s.Paging)
}

func TestLookupConstruction(t *testing.T) {
	l := query.Lookup{ID: "order-42"}
	assert.Equal(t, "order-42", l.ID)
}
