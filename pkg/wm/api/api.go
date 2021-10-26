package api

import (
	"fmt"
	"net/http"

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

		r.Get("/models/{"+paramModelID+"}/runs", a.wh(a.getModelRuns))
		r.Get("/models/{"+paramModelID+"}/parameters", a.wh(a.getModelParameters))
		r.Get("/datacubes", a.wh(a.getDatacubes))
		r.Get("/datacubes/count", a.wh(a.countDatacubes))
		r.Get("/indicator-data", a.wh(a.getIndicatorData))
		r.Get("/concepts", a.wh(a.getConcepts))
		//New endpoints for data in Minio
		r.Get("/output/hierarchy", a.wh(a.getDataOutputHierarchy))
		r.Get("/output/hierarchy-lists", a.wh(a.getDataOutputRegionLists))
		r.Get("/output/timeseries", a.wh(a.getDataOutputTimeseries))
		r.Get("/output/stats", a.wh(a.getDataOutputStats))
		r.Get("/output/regional-data", a.wh(a.getDataOutputRegional))
		r.Get("/output/regional-stats", a.wh(a.getRegionalDataOutputStats))
		r.Get("/output/raw-data", a.wh(a.getDataOutputRaw))
		r.Get("/output/qualifier-timeseries", a.wh(a.getDataOutputQualifierTimeseries))
		r.Get("/output/qualifier-data", a.wh(a.getDataOutputQualifierData))
	})

	r.Route("/maas/output/tiles", func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}", paramZoom, paramX, paramY), a.wh(a.getTile))
	})

	// TODO: Merge grid tiles route (/maas/output/tiles) with this route
	r.Route("/maas/tiles", func(r chi.Router) {
		r.Get(fmt.Sprintf("/grid-output/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}", paramZoom, paramX, paramY), a.wh(a.getTile))
		r.Get(fmt.Sprintf("/{%s}/{%s:[0-9]+}/{%s:[0-9]+}/{%s:[0-9]+}", paramTileSetName, paramZoom, paramX, paramY), a.wh(a.getVectorTile))
	})

	return r, nil
}

// wh returns wrapped handler function that wraps given handler. Wrapped handler calls given handler and handles error.
func (a *api) wh(handler func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		// Handle error
		errCode := wm.ErrorCode(err)
		if errCode == wm.EINTERNAL {
			a.logger.Errorf("ERROR: %s\n", err)
		}

		status := http.StatusInternalServerError

		switch errCode {
		case wm.ENOTFOUND:
			status = http.StatusNotFound
		case wm.EINVALID:
			status = http.StatusBadRequest
		case wm.ECONFLICT:
			status = http.StatusConflict
		case wm.EINTERNAL:
			status = http.StatusInternalServerError
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(wm.ErrorMessage(err)))
	}
}
