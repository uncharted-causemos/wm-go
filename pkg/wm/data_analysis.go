package wm

import "time"

// Analysis represent an analysis object
type Analysis struct {
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	State       AnalysisState `json:"state"`
	ModifiedAt  time.Time     `json:"modified_at"`
	CreatedAt   time.Time     `json:"created_at"`
}

// AnalysisState represent arbitrary state object for the analysis
type AnalysisState map[string]interface{}

/**
ES mapping for analysis index
{
	"mappings": {
		"properties": {
				"id": { "type": "keyword" },
				"project_id": { "type": "keyword" },
				"title": { "type": "text" },
				"description": { "type": "text" },
				"state": { "type": "object", "enabled": false },
				"modified_at": { "type": "date" },
				"created_at": { "type": "date" }
		}
	}
}
*/

// DataAnalysis defines the methods that data analysis database implementation needs to satisfy.
type DataAnalysis interface {
	GetAnalysisByID(AnalysisID string) (*Analysis, error)

	GetAnalyses(filters []*Filter) ([]*Analysis, error)

	CreateAnalysis(payload *Analysis) (*Analysis, error)

	UpdateAnalysis(analysisID string, payload *Analysis) (*Analysis, error)

	DeleteAnalysis(analysisID string) error

	UpdateAnalysisState(analysisID string, state AnalysisState) (AnalysisState, error)
}
