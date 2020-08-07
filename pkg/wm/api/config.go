package api

import (
	"errors"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	"go.uber.org/zap"
)

// Config defines the parameters needed to instantiate the API router.
type Config struct {
	Graph         wm.Graph
	KnowledgeBase wm.KnowledgeBase
	MaaS          wm.MaaS
	MaaSStorage   wm.MaaSStorage
	Logger        *zap.SugaredLogger
}

// init validates the config and fills in defaults for missing optional
// parameters.
func (cfg *Config) init() error {
	/*
		if cfg.Graph == nil {
			return errors.New("Graph cannot be nil")
		}
	*/
	if cfg.KnowledgeBase == nil {
		return errors.New("KnowledgeBase cannot be nil")
	}
	if cfg.MaaS == nil {
		return errors.New("MaaS cannot be nil")
	}
	if cfg.Logger == nil {
		return errors.New("Logger cannot be nil")
	}
	return nil
}
