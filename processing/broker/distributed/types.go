package distributed

type DistributionState struct {
	Subscribers   []SubscriberInfo
	Subscriptions []SubscriptionInfo
	EventSources  []string
}

type SubscriberInfo struct {
	ID      string
	GroupID string
	NodeID  string
	Active  bool
}

type SubscriptionInfo struct {
	ID           string
	SubscriberID string
	EventSources []string
}

type ChangeAction int

const (
	ChangeActionAdd ChangeAction = iota
	ChangeActionRemove
	ChangeActionReplace
)

type SubscriptionStateChange struct {
	Action       ChangeAction
	SubscriberID string
	EventSources []string
}
