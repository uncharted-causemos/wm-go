package modelservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const timeout = 10 * time.Second

// Service connects to and communicates with the model rest api service (https://model-service.worldmodelers.com/ui)
type Service struct {
	client   *http.Client
	url      string
	username string
	password string
}

// New instantiates and returns a new KB using the provided Config.
func New(url, username, password string) (*Service, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	return &Service{client, url, username, password}, nil
}

// GetModelParameters returns model parameters
func (s *Service) GetModelParameters(model string) ([]*wm.ModelParameter, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/model_parameters/%s", s.url, model), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parameters []*wm.ModelParameter
	if err := json.Unmarshal(body, &parameters); err != nil {
		return nil, err
	}
	return parameters, nil
}

// GetConcepts returns concept names
func (s *Service) GetConcepts() ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/list_concepts", s.url), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var concepts []string
	if err := json.Unmarshal(body, &concepts); err != nil {
		return nil, err
	}
	return concepts, nil
}
