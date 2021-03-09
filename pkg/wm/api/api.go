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
	paramProjectID   = "projectID"
	paramZoom        = "zoom"
	paramX           = "x"
	paramY           = "y"
	paramModelID     = "modelID"
	paramRunID       = "runID"
	paramTileSetName = "tileSetName"
)

type api struct {
	maas       wm.MaaS
	dataOutput wm.DataOutput
	vectorTile wm.VectorTile
	logger     *zap.SugaredLogger
}

// New returns a chi router with the various endpoints defined.
func New(cfg *Config) (chi.Router, error) {
	if err := cfg.init(); err != nil {
		return nil, err
	}

	a := api{
		maas:       cfg.MaaS,
		dataOutput: cfg.DataOutput,
		vectorTile: cfg.VectorTile,
		logger:     cfg.Logger,
	}

	r := chi.NewRouter()

	r.Route("/maas", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		r.Get("/models/{"+paramModelID+"}/runs", a.getModelRuns)
		r.Get("/models/{"+paramModelID+"}/parameters", a.getModelParameters)
		r.Get("/datacubes", a.getDatacubes)
		r.Get("/datacubes/count", a.countDatacubes)
		r.Get("/indicator-data", a.getIndicatorData)
		r.Get("/concepts", a.getConcepts)
		r.Get("/output/{"+paramRunID+"}/timeseries", a.getModelOutputTimeseries)
		r.Get("/output/{"+paramRunID+"}/stats", a.getModelOutputStats)
	})

	r.Route("/maas/output/tiles", func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}", paramZoom, paramX, paramY), a.getTile)
	})

	r.Route("/maas/tiles", func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s}/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}.pbf", paramTileSetName, paramZoom, paramX, paramY), a.getVectorTile)
	})

	return r, nil
}
