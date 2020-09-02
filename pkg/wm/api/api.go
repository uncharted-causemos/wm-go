package api

import (
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	"go.uber.org/zap"
)

// URL parameter strings
const (
	paramProjectID = "projectID"
	paramZoom      = "zoom"
	paramX         = "x"
	paramY         = "y"
	paramModelID   = "modelID"
)

type api struct {
	graph       wm.Graph
	kb          wm.KnowledgeBase
	maas        wm.MaaS
	maasStorage wm.MaaSStorage
	logger      *zap.SugaredLogger
}

// New returns a chi router with the various endpoints defined.
func New(cfg *Config) (chi.Router, error) {
	if err := cfg.init(); err != nil {
		return nil, err
	}

	a := api{
		graph:       cfg.Graph,
		kb:          cfg.KnowledgeBase,
		maas:        cfg.MaaS,
		maasStorage: cfg.MaaSStorage,
		logger:      cfg.Logger,
	}

	r := chi.NewRouter()

	r.Route("/{"+paramProjectID+":[a-f0-9-]+}", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		r.Get("/facets", a.getFacets)
	})

	r.Route("/maas", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		r.Get("/models/{"+paramModelID+"}/runs", a.getModelRuns)
		r.Get("/models/{"+paramModelID+"}/parameters", a.getModelParameters)
		r.Get("/datacubes", a.getDatacubes)
		r.Get("/datacubes/count", a.countDatacubes)
		r.Get("/concepts", a.getConcepts)
	})

	r.Route("/maas/output/tiles", func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}", paramZoom, paramX, paramY), a.getTile)
	})

	return r, nil
}
