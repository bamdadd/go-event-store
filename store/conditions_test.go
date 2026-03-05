package store_test

import (
	"testing"

	"github.com/bamdadd/go-event-store/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int { return &i }

func TestNoConditionAlwaysPasses(t *testing.T) {
	ok, err := store.EvaluateCondition(store.NoCondition(), intPtr(5))
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestNoConditionPassesWithNilPosition(t *testing.T) {
	ok, err := store.EvaluateCondition(store.NoCondition(), nil)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestPositionIsMatchesExactPosition(t *testing.T) {
	ok, err := store.EvaluateCondition(store.PositionIs(intPtr(3)), intPtr(3))
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestPositionIsFailsOnMismatch(t *testing.T) {
	ok, err := store.EvaluateCondition(store.PositionIs(intPtr(3)), intPtr(5))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPositionIsNilMatchesNilPosition(t *testing.T) {
	ok, err := store.EvaluateCondition(store.PositionIs(nil), nil)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestPositionIsNilFailsWhenPositionExists(t *testing.T) {
	ok, err := store.EvaluateCondition(store.PositionIs(nil), intPtr(0))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestStreamIsEmptyPassesWhenNilPosition(t *testing.T) {
	ok, err := store.EvaluateCondition(store.StreamIsEmpty(), nil)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestStreamIsEmptyFailsWhenPositionExists(t *testing.T) {
	ok, err := store.EvaluateCondition(store.StreamIsEmpty(), intPtr(0))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestAndPassesWhenAllPass(t *testing.T) {
	ok, err := store.EvaluateCondition(
		store.And(store.PositionIs(intPtr(3)), store.NoCondition()),
		intPtr(3),
	)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAndFailsWhenAnyFails(t *testing.T) {
	ok, err := store.EvaluateCondition(
		store.And(store.PositionIs(intPtr(3)), store.StreamIsEmpty()),
		intPtr(3),
	)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestOrPassesIfAnyPasses(t *testing.T) {
	ok, err := store.EvaluateCondition(
		store.Or(store.StreamIsEmpty(), store.PositionIs(intPtr(5))),
		intPtr(5),
	)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestOrFailsIfAllFail(t *testing.T) {
	ok, err := store.EvaluateCondition(
		store.Or(store.StreamIsEmpty(), store.PositionIs(intPtr(3))),
		intPtr(5),
	)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestNestedAndOrComposition(t *testing.T) {
	cond := store.Or(
		store.And(store.PositionIs(intPtr(5)), store.NoCondition()),
		store.StreamIsEmpty(),
	)
	ok, err := store.EvaluateCondition(cond, intPtr(5))
	require.NoError(t, err)
	assert.True(t, ok)
}
