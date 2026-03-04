package projection

import "encoding/json"

type Projection struct {
	ID       string
	Name     string
	Source   json.RawMessage
	State    json.RawMessage
	Metadata json.RawMessage
}
