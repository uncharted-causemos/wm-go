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
	// Document facets
	Location        []Facet `json:"location,omitempty"`
	Organization    []Facet `json:"organization,omitempty"`
	PublicationYear []Facet `json:"publicationYear,omitempty"`

	// Factor facets
	Concept        []Facet `json:"concept,omitempty"`
	GroundingScore []Facet `json:"groundingScore,omitempty"`

	// Statement facets
	Cause            []Facet `json:"cause,omitempty"`
	Effect           []Facet `json:"effect,omitempty"`
	Polarity         []Facet `json:"polarity,omitempty"`
	BeliefScore      []Facet `json:"beliefScore,omitempty"`
	NumEvidence      []Facet `json:"numEvidence,omitempty"`
	Reader           []Facet `json:"reader,omitempty"`
	RefutingEvidence []Facet `json:"refutingEvidence,omitempty"`
	Quality          []Facet `json:"quality,omitempty"`
	Hedging          []Facet `json:"hedging,omitempty"`
	EvidenceSource   []Facet `json:"evidenceSource,omitempty"`
}
