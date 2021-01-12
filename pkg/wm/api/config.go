package api

import (
	"errors"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	"go.uber.org/zap"
)

// Config defines the parameters needed to instantiate the API router.
type Config struct {
	MaaS           wm.MaaS
	DataOutputTile wm.DataOutputTile
	Logger         *zap.SugaredLogger
}

// init validates the config and fills in defaults for missing optional
// parameters.
func (cfg *Config) init() error {
	if cfg.MaaS == nil {
		return errors.New("MaaS cannot be nil")
	}
	if cfg.Logger == nil {
		return errors.New("Logger cannot be nil")
	}
	return nil
}
