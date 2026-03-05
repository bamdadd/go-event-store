package postgres

import (
	"fmt"
	"strings"

	"github.com/bamdadd/go-event-store/store"
	"github.com/bamdadd/go-event-store/types"
)

func buildScanQuery(target types.Targetable, cfg store.ScanConfig) (string, []any) {
	var conditions []string
	var args []any
	argIdx := 1

	switch t := target.(type) {
	case types.StreamIdentifier:
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, t.Category)
		argIdx++
		conditions = append(conditions, fmt.Sprintf("stream = $%d", argIdx))
		args = append(args, t.Stream)
		argIdx++
	case types.CategoryIdentifier:
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, t.Category)
		argIdx++
	case types.LogIdentifier:
		// no filter
	}

	for _, c := range cfg.Constraints {
		switch qc := c.(type) {
		case store.SequenceNumberAfterConstraint:
			conditions = append(conditions, fmt.Sprintf("sequence_number > $%d", argIdx))
			args = append(args, qc.SequenceNumber)
			argIdx++
		}
	}

	query := "SELECT id, name, stream, category, position, sequence_number, payload, observed_at, occurred_at FROM events"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY sequence_number ASC"
	return query, args
}

func buildLatestQuery(target types.Targetable) (string, []any) {
	var conditions []string
	var args []any
	argIdx := 1

	switch t := target.(type) {
	case types.StreamIdentifier:
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, t.Category)
		argIdx++
		conditions = append(conditions, fmt.Sprintf("stream = $%d", argIdx))
		args = append(args, t.Stream)
		argIdx++
	case types.CategoryIdentifier:
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, t.Category)
		argIdx++
	case types.LogIdentifier:
		// no filter
	}

	query := "SELECT id, name, stream, category, position, sequence_number, payload, observed_at, occurred_at FROM events"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY sequence_number DESC LIMIT 1"
	return query, args
}
