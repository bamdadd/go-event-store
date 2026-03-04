package distributed

func computeChanges(state DistributionState) []SubscriptionStateChange {
	var changes []SubscriptionStateChange

	activeSubscribers := make([]SubscriberInfo, 0)
	for _, s := range state.Subscribers {
		if s.Active {
			activeSubscribers = append(activeSubscribers, s)
		}
	}

	activeIDs := make(map[string]bool)
	for _, s := range activeSubscribers {
		activeIDs[s.ID] = true
	}
	for _, sub := range state.Subscriptions {
		if !activeIDs[sub.SubscriberID] {
			changes = append(changes, SubscriptionStateChange{
				Action:       ChangeActionRemove,
				SubscriberID: sub.SubscriberID,
			})
		}
	}

	if len(activeSubscribers) == 0 {
		return changes
	}

	chunkSize := len(state.EventSources) / len(activeSubscribers)
	remainder := len(state.EventSources) % len(activeSubscribers)

	offset := 0
	for i, sub := range activeSubscribers {
		size := chunkSize
		if i < remainder {
			size++
		}
		sources := state.EventSources[offset : offset+size]
		offset += size

		changes = append(changes, SubscriptionStateChange{
			Action:       ChangeActionReplace,
			SubscriberID: sub.ID,
			EventSources: sources,
		})
	}

	return changes
}
