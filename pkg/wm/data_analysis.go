package wm

import "time"

// Analysis represent an analysis object
type Analysis struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	ModifiedAt  time.Time `json:"modified_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// DataAnalysis defines the methods that data analysis database implementation needs to satisfy.
type DataAnalysis interface {
	GetAnalyses(filters []*Filter) ([]*Analysis, error)

	CreateAnalysis(payload *Analysis) (*Analysis, error)

	// UpdateAnalysis(AnalysisID string, payload *Analysis) (*Analysis, error)

	// DeleteAnalysis(analysisID string) error

	// UpdateAnalysisState(analysisID string, state string) error
}
