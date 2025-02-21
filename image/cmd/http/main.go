package main

import (
	"context"
	"log"
	"net/http"

	commonenv "smapp/common/env"
	commonmw "smapp/common/middleware"
	"smapp/image/handlers"
	"smapp/image/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/mux"
)

func main() {
	defaultTimeout, err := commonenv.GetEnvDuration("DEFAULT_TIMEOUT")
	if err != nil {
		log.Fatal(err)
	}
	profileImgLimit, err := commonenv.GetEnvInt64("PROFILE_IMG_LIMIT")
	if err != nil {
		log.Fatal(err)
	}
	postImgLimit, err := commonenv.GetEnvInt64("POST_IMG_LIMIT")
	if err != nil {
		log.Fatal(err)
	}
	policyTTL, err := commonenv.GetEnvDuration("POLICY_TTL")
	if err != nil {
		log.Fatal(err)
	}
	bucket, err := commonenv.GetEnv("S3_BUCKET")
	if err != nil {
		log.Fatal(err)
	}
	region, err := commonenv.GetEnv("S3_REGION")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	generateUploadFormService := service.NewGenerateUploadForm(cfg, policyTTL, bucket, region)

	r := mux.NewRouter()
	r.Handle(
		"/upload-form/profile",
		commonmw.ParseUserID(handlers.GenerateUploadForm(generateUploadFormService, "profile", profileImgLimit)),
	).Methods(http.MethodGet)
	r.Handle(
		"/upload-form/post",
		commonmw.ParseUserID(handlers.GenerateUploadForm(generateUploadFormService, "post", postImgLimit)),
	).Methods(http.MethodGet)

	r.Use(commonmw.WithRequestContextTimeout(defaultTimeout))

	srv := &http.Server{
		Addr:        ":8080",
		Handler:     r,
		ReadTimeout: defaultTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
