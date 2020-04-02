package main

import (
	"compress/flate"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/api"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/elastic"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/env"
	"go.uber.org/zap"
)

func main() {
	s, err := env.Load("wm.env")
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
	r.Use(render.SetContentType(render.ContentTypeJSON))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})
	r.Use(c.Handler)

	es, err := elastic.New(&elastic.Config{
		Addr: s.ElasticURL,
	})

	apiRouter, err := api.New(&api.Config{
		KnowledgeBase: es,
		MaaS:          es,
		Logger:        sugar,
	})
	if err != nil {
		sugar.Fatal(err)
	}

	r.Mount("/", apiRouter)

	sugar.Infof("Listening on %s", s.Addr)
	sugar.Fatal(http.ListenAndServe(s.Addr, r))
}
