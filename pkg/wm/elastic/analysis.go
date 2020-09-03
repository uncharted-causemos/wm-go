package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const analysisIndex = "analysis"

// indexAnalysis index given analysis
func (es *ES) indexAnalysis(analysis *wm.Analysis) (*wm.Analysis, error) {
	body, err := json.Marshal(analysis)
	if err != nil {
		return nil, err
	}
	res, err := es.client.Index(
		analysisIndex,
		bytes.NewReader(body),
		es.client.Index.WithDocumentID(analysis.ID),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, errors.New(read(res.Body))
	}
	return analysis, nil
}

// GetAnalysisByID Gets an analysis by ID
func (es *ES) GetAnalysisByID(analysisID string) (*wm.Analysis, error) {
	var analysis *wm.Analysis
	res, err := es.client.GetSource(analysisIndex, analysisID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	resBody := read(res.Body)
	if res.IsError() {
		return nil, errors.New(resBody)
	}
	if err := json.Unmarshal([]byte(resBody), &analysis); err != nil {
		return nil, err
	}
	return analysis, nil
}

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
	return es.indexAnalysis(newAnalysis)
}

// UpdateAnalysis updates the analysis with given ID. Creates new one with given ID if not already exist
func (es *ES) UpdateAnalysis(analysisID string, payload *wm.Analysis) (*wm.Analysis, error) {
	analysis, err := es.GetAnalysisByID(analysisID)
	if err != nil {
		return nil, err
	}
	if analysis == nil {
		analysis = &wm.Analysis{
			ID:        analysisID,
			CreatedAt: time.Now(),
		}
	}
	analysis.Title = payload.Title
	analysis.Description = payload.Description
	analysis.ModifiedAt = time.Now()
	return es.indexAnalysis(analysis)
}

// DeleteAnalysis deletes the analysis with given ID
func (es *ES) DeleteAnalysis(analysisID string) error {
	res, err := es.client.Delete(analysisIndex, analysisID)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return errors.New(read(res.Body))
	}
	return nil
}
