package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
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
		es.client.Index.WithRefresh("true"),
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
	// TODO: Implement proper pagination with size, from and sort options
	body := map[string]interface{}{
		"size": 100,
		"sort": []map[string]string{
			{"modified_at": "desc"},
		},
	}
	if len(query) > 0 {
		body["query"] = query
	}
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	res, err := es.client.Search(
		es.client.Search.WithIndex(analysisIndex),
		es.client.Search.WithBody(&buf),
	)
	defer res.Body.Close()
	resBody := read(res.Body)
	if res.IsError() {
		return nil, fmt.Errorf("ES response error: %s", resBody)
	}
	hits := gjson.Get(resBody, "hits.hits").Array()
	for _, hit := range hits {
		doc := hit.Get("_source").String()
		var analysis *wm.Analysis
		if err := json.Unmarshal([]byte(doc), &analysis); err != nil {
			return nil, err
		}
		analyses = append(analyses, analysis)
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

// UpdateAnalysis updates the analysis with given ID.
func (es *ES) UpdateAnalysis(analysisID string, payload *wm.Analysis) (*wm.Analysis, error) {
	analysis, err := es.GetAnalysisByID(analysisID)
	if err != nil {
		return nil, err
	}
	if analysis == nil {
		return nil, fmt.Errorf("Analysis with given ID not found: %s", analysisID)
	}
	analysis.Title = payload.Title
	analysis.Description = payload.Description
	analysis.ModifiedAt = time.Now()
	return es.indexAnalysis(analysis)
}

// UpdateAnalysisState updates the state of the analysis
func (es *ES) UpdateAnalysisState(analysisID string, state wm.AnalysisState) (wm.AnalysisState, error) {
	body := map[string]interface{}{
		"doc": map[string]wm.AnalysisState{
			"state": state,
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	res, err := es.client.Update(analysisIndex, analysisID, &buf)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("ES response error: %s", read(res.Body))
	}
	return state, nil
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
