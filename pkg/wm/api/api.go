package api

import (
	"fmt"

	"github.com/go-chi/chi"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	"go.uber.org/zap"
)

// URL parameter strings
const (
	paramProjectID = "projectID"
	paramZoom      = "zoom"
	paramX         = "x"
	paramY         = "y"
)

type api struct {
	graph  wm.Graph
	kb     wm.KnowledgeBase
	maas   wm.MaaS
	logger *zap.SugaredLogger
}

// New returns a chi router with the various endpoints defined.
func New(cfg *Config) (chi.Router, error) {
	if err := cfg.init(); err != nil {
		return nil, err
	}

	a := api{
		graph:  cfg.Graph,
		kb:     cfg.KnowledgeBase,
		maas:   cfg.MaaS,
		logger: cfg.Logger,
	}

	r := chi.NewRouter()

	r.Get("/{"+paramProjectID+":[a-f0-9-]+}/facets", a.getFacets)
	r.Get(fmt.Sprintf("/tiles/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}", paramZoom, paramX, paramY), a.getTile)

	return r, nil
}
