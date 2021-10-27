package main

import (
	"compress/flate"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	mw "gitlab.uncharted.software/WM/wm-go/pkg/middleware"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/api"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/elastic"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm/env"
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
		dev := zap.NewDevelopmentConfig()
		dev.DisableStacktrace = true
		logger, err = dev.Build()
	case "prod":
		prod := zap.NewProductionConfig()
		prod.DisableStacktrace = true
		logger, err = prod.Build()
	default:
		err = fmt.Errorf("invalid 'mode' flag: %s", s.Mode)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	r := chi.NewRouter()

	color := true
	if s.Mode == "prod" {
		color = false
	}
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.NewZapRequestLogger(logger, color))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.DefaultCompression))

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
	if err != nil {
		sugar.Fatal(err)
	}

	s3, err := Storage.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(s.AwsS3Id, s.AwsS3Secret, s.AwsS3Token),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(endpoints.UsEast1RegionID),
		Endpoint:         aws.String(s.AwsS3URL), // LocalStack/Minio S3 Port
	}, sugar)
	if err != nil {
		sugar.Fatal(err)
	}

	apiRouter, err := api.New(&api.Config{
		MaaS:       es,
		DataOutput: s3,
		VectorTile: s3,
		Logger:     sugar,
	})
	if err != nil {
		sugar.Fatal(err)
	}

	r.Mount("/", apiRouter)

	sugar.Infof("Listening on %s", s.Addr)
	sugar.Fatal(http.ListenAndServe(s.Addr, r))
}
