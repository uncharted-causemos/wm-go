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
	dataOutput wm.DataOutput
	vectorTile wm.VectorTile
	logger     *zap.SugaredLogger
}

// New returns a chi router with the various endpoints defined.
func New(cfg *Config) (chi.Router, error) {
	op := "api.New"
	if err := cfg.init(); err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	a := api{
		dataOutput: cfg.DataOutput,
		vectorTile: cfg.VectorTile,
		logger:     cfg.Logger,
	}

	r := chi.NewRouter()

	r.Route("/maas", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		//Endpoints for data stored in Minio
		r.Get("/output/region-lists", a.wh(a.getDataOutputRegionLists))
		r.Get("/output/qualifier-counts", a.wh(a.getDataOutputQualifierCounts))
		r.Get("/output/qualifier-lists", a.wh(a.getDataOutputQualifierLists))
		r.Get("/output/timeseries", a.wh(a.getDataOutputTimeseries))
		r.Get("/output/extrema", a.wh(a.getDataOutputExtrema))
		r.Get("/output/sparkline", a.wh(a.getDataOutputSparkline))
		r.Post("/output/bulk-timeseries/regions", a.wh(a.getBulkDataOutputRegionTimeseries))
		r.Post("/output/bulk-timeseries/generic", a.wh(a.getBulkDataOutputGenericTimeseries))
		r.Post("/output/aggregate-timeseries", a.wh(a.getAggregateDataOutputTimeseries))
		r.Get("/output/stats", a.wh(a.getDataOutputStats))
		r.Get("/output/regional-data", a.wh(a.getDataOutputRegional))
		r.Post("/output/bulk-regional-data", a.wh(a.getBulkDataOutputRegional))
		r.Get("/output/regional-stats", a.wh(a.getRegionalDataOutputStats))
		r.Get("/output/raw-data", a.wh(a.getDataOutputRaw))
		r.Get("/output/qualifier-timeseries", a.wh(a.getDataOutputQualifierTimeseries))
		r.Get("/output/qualifier-data", a.wh(a.getDataOutputQualifierData))
		r.Get("/output/qualifier-regional", a.wh(a.getDataOutputQualifierRegional))
		r.Get("/output/pipeline-results", a.wh(a.getDataOutputPipelineResults))
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
		errMessage := wm.ErrorMessage(err)

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
		w.Write([]byte(errMessage))
		if errCode == wm.EINTERNAL {
			// Log error if it's an internal server error
			a.logger.Error(err)
		} else {
			a.logger.Debug(err)
		}
	}
}
