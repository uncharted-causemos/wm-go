package api

import (
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	"go.uber.org/zap"
)

// Config defines the parameters needed to instantiate the API router.
type Config struct {
	MaaS       wm.MaaS
	DataOutput wm.DataOutput
	VectorTile wm.VectorTile
	Logger     *zap.SugaredLogger
}

// init validates the config and fills in defaults for missing optional
// parameters.
func (cfg *Config) init() error {
	op := "Config.init"
	if cfg.MaaS == nil {
		return &wm.Error{Op: op, Message: "MaaS cannot be nil"}
	}
	if cfg.Logger == nil {
		return &wm.Error{Op: op, Message: "Logger cannot be nil"}
	}
	if cfg.DataOutput == nil {
		return &wm.Error{Op: op, Message: "DataOutput cannot be nil"}
	}
	if cfg.VectorTile == nil {
		return &wm.Error{Op: op, Message: "Logger cannot be nil"}
	}
	return nil
}
