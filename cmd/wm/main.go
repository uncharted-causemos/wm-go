package main

import (
	"compress/flate"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/api"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/dgraph"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/elastic"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/env"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/modelservice"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/storage"
	"go.uber.org/zap"
)

const envFile = "wm.env"

func main() {
	s, err := env.Load(envFile)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the logger
	var logger *zap.Logger
	switch s.Mode {
	case "dev":
		logger, err = zap.NewDevelopment()
	case "prod":
		logger, err = zap.NewProduction()
	default:
		err = fmt.Errorf("Invalid 'mode' flag: %s", s.Mode)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.DefaultCompression))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})
	r.Use(c.Handler)

	ms, err := modelservice.New(s.MaasURL, s.MaasUser, s.MaasPassword)
	if err != nil {
		sugar.Fatal(err)
	}
	es, err := elastic.New(&elastic.Config{
		Addr:         s.ElasticURL,
		ModelService: ms,
	})
	if err != nil {
		sugar.Fatal(err)
	}

	s3, err := storage.New(nil, "")
	if err != nil {
		sugar.Fatal(err)
	}

	dg, err := dgraph.New(&dgraph.Config{
		Addrs: s.DgraphURLS,
	})
	if err != nil {
		sugar.Fatal(err)
	}

	apiRouter, err := api.New(&api.Config{
		KnowledgeBase: es,
		MaaS:          es,
		MaaSStorage:   s3,
		Graph:         dg,
		Logger:        sugar,
	})
	if err != nil {
		sugar.Fatal(err)
	}

	r.Mount("/", apiRouter)

	sugar.Infof("Listening on %s", s.Addr)
	sugar.Fatal(http.ListenAndServe(s.Addr, r))
}
