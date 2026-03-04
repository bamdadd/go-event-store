package distributed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeChangesNoSubscribers(t *testing.T) {
	state := DistributionState{
		EventSources: []string{"src-1", "src-2"},
	}
	changes := computeChanges(state)
	assert.Empty(t, changes)
}

func TestComputeChangesSingleSubscriber(t *testing.T) {
	state := DistributionState{
		Subscribers:  []SubscriberInfo{{ID: "s1", Active: true}},
		EventSources: []string{"src-1", "src-2", "src-3"},
	}
	changes := computeChanges(state)
	assert.Len(t, changes, 1)
	assert.Equal(t, ChangeActionReplace, changes[0].Action)
	assert.Equal(t, "s1", changes[0].SubscriberID)
	assert.Equal(t, []string{"src-1", "src-2", "src-3"}, changes[0].EventSources)
}

func TestComputeChangesEvenDistribution(t *testing.T) {
	state := DistributionState{
		Subscribers: []SubscriberInfo{
			{ID: "s1", Active: true},
			{ID: "s2", Active: true},
		},
		EventSources: []string{"src-1", "src-2", "src-3", "src-4"},
	}
	changes := computeChanges(state)
	assert.Len(t, changes, 2)
	assert.Len(t, changes[0].EventSources, 2)
	assert.Len(t, changes[1].EventSources, 2)
}

func TestComputeChangesUnevenDistribution(t *testing.T) {
	state := DistributionState{
		Subscribers: []SubscriberInfo{
			{ID: "s1", Active: true},
			{ID: "s2", Active: true},
		},
		EventSources: []string{"src-1", "src-2", "src-3"},
	}
	changes := computeChanges(state)
	assert.Len(t, changes, 2)
	assert.Len(t, changes[0].EventSources, 2)
	assert.Len(t, changes[1].EventSources, 1)
}

func TestComputeChangesRemovesInactiveSubscribers(t *testing.T) {
	state := DistributionState{
		Subscribers: []SubscriberInfo{
			{ID: "s1", Active: true},
			{ID: "s2", Active: false},
		},
		Subscriptions: []SubscriptionInfo{
			{ID: "sub-1", SubscriberID: "s2", EventSources: []string{"src-1"}},
		},
		EventSources: []string{"src-1", "src-2"},
	}
	changes := computeChanges(state)

	var removeChanges []SubscriptionStateChange
	var replaceChanges []SubscriptionStateChange
	for _, c := range changes {
		if c.Action == ChangeActionRemove {
			removeChanges = append(removeChanges, c)
		} else {
			replaceChanges = append(replaceChanges, c)
		}
	}
	assert.Len(t, removeChanges, 1)
	assert.Equal(t, "s2", removeChanges[0].SubscriberID)
	assert.Len(t, replaceChanges, 1)
	assert.Equal(t, "s1", replaceChanges[0].SubscriberID)
	assert.Len(t, replaceChanges[0].EventSources, 2)
}

func TestComputeChangesNoSourcesAssignsEmpty(t *testing.T) {
	state := DistributionState{
		Subscribers:  []SubscriberInfo{{ID: "s1", Active: true}},
		EventSources: []string{},
	}
	changes := computeChanges(state)
	assert.Len(t, changes, 1)
	assert.Empty(t, changes[0].EventSources)
}
