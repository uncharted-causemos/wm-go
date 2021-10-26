package api

import (
	"net/http"

	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type datacubesResponse struct {
	*wm.Datacube
}

type countDatacubesResponse uint64

type indicatorDataResponse struct {
	*wm.IndicatorDataPoint
}

// Render allows to satisfy the render.Renderer interface.
func (d *datacubesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render satisfies the render.Renderer interface.
func (cd countDatacubesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render allows to satisfy the render.Renderer interface.
func (d *indicatorDataResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getDatacubes(w http.ResponseWriter, r *http.Request) error {
	filters, err := getFilters(r, wm.ContextDatacube)
	if err != nil {
		return err
	}
	datacubes, err := a.maas.SearchDatacubes(filters)
	if err != nil {
		return err
	}
	list := []render.Renderer{}
	for _, datacube := range datacubes {
		list = append(list, &datacubesResponse{datacube})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) countDatacubes(w http.ResponseWriter, r *http.Request) error {
	filters, err := getFilters(r, wm.ContextDatacube)
	if err != nil {
		return err
	}
	count, err := a.maas.CountDatacubes(filters)
	if err != nil {
		return err
	}
	render.Render(w, r, countDatacubesResponse(count))
	return nil
}

func (a *api) getIndicatorData(w http.ResponseWriter, r *http.Request) error {
	indicator := getIndicator(r)
	model := getModel(r)
	units, ok := getUnits(r)
	if !ok {
		units = []string{}
	}

	//For now, only handle a single indicatorName
	indicatorData, err := a.maas.GetIndicatorData(indicator, model, units)
	if err != nil {
		return err
	}
	list := []render.Renderer{}
	for _, indicator := range indicatorData {
		list = append(list, &indicatorDataResponse{indicator})
	}
	render.RenderList(w, r, list)
	return nil
}
