package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const analysisIndex = "analysis"

// GetAnalyses returns a list of analysis that meets given filter criteria
func (es *ES) GetAnalyses(filters []*wm.Filter) ([]*wm.Analysis, error) {
	var analyses []*wm.Analysis
	options := queryOptions{
		filters: filters,
	}
	query, err := buildQuery(options)
	body := map[string]interface{}{
		"size": defaultSize,
	}
	if len(query) > 0 {
		body["query"] = query
	}
	if err != nil {
		return nil, err
	}
	return analyses, nil
}

// CreateAnalysis creates an analysis
func (es *ES) CreateAnalysis(payload *wm.Analysis) (*wm.Analysis, error) {
	newAnalysis := &wm.Analysis{
		ID:          uuid.New().String(),
		ProjectID:   payload.ProjectID,
		Title:       payload.Title,
		Description: payload.Description,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
	}
	body, err := json.Marshal(newAnalysis)
	if err != nil {
		return nil, err
	}
	res, err := es.client.Index(
		analysisIndex,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, errors.New(read(res.Body))
	}
	return newAnalysis, nil
}
