package elastic

import (
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// GetModelParameters returns model runs
func (es *ES) GetModelParameters(model string) ([]*wm.ModelParameter, error) {
	// TODO: Get this from ES instead of old maas api when Galois ES is up.
	return es.modelService.GetModelParameters(model)
}
