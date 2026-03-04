package types_test

import (
	"testing"

	"github.com/logicblocks/event-store/types"
	"github.com/stretchr/testify/assert"
)

func TestStreamIdentifierIsTargetable(t *testing.T) {
	var target types.Targetable = types.StreamIdentifier{Category: "orders", Stream: "order-123"}
	assert.NotNil(t, target)
}

func TestCategoryIdentifierIsTargetable(t *testing.T) {
	var target types.Targetable = types.CategoryIdentifier{Category: "orders"}
	assert.NotNil(t, target)
}

func TestLogIdentifierIsTargetable(t *testing.T) {
	var target types.Targetable = types.LogIdentifier{}
	assert.NotNil(t, target)
}

func TestStreamIdentifierFieldValues(t *testing.T) {
	id := types.StreamIdentifier{Category: "orders", Stream: "order-42"}
	assert.Equal(t, "orders", id.Category)
	assert.Equal(t, "order-42", id.Stream)
}
