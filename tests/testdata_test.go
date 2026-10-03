package toon_test

type profile struct {
	ID     int     `toon:"id"`
	Name   string  `toon:"name"`
	Active bool    `toon:"active"`
	Email  *string `toon:"email,omitempty"`
}

type usersPayload struct {
	Users []profile `toon:"users"`
	Count int       `toon:"count"`
}

type metricEvent struct {
	Type   string `toon:"type"`
	Values []int  `toon:"values"`
}

type typedEnvelope struct {
	Events []metricEvent `toon:"events"`
}
