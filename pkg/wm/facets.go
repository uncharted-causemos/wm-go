package wm

// Facet is an individual bucket in a facets response.
type Facet struct {
	ID         interface{} `json:"id"`
	Count      int         `json:"count"`
	StartValue float64     `json:"startValue,omitempty"`
	EndValue   float64     `json:"endValue,omitempty"`
}

// Facets represent the results of a facets query.
type Facets struct {
	// Datacubes facets
	Type     []Facet `json:"type,omitempty"`
	Category []Facet `json:"category,omitempty"`
	// TODO: add more fields
}
