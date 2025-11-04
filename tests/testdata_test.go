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

type mixedEnvelope struct {
	Events []any `toon:"events"`
}

type typedEnvelope struct {
	Events []metricEvent `toon:"events"`
}

type bucket struct {
	Values []int  `toon:"values"`
	Label  string `toon:"label"`
}

type bucketSet struct {
	Buckets []bucket `toon:"buckets"`
}
