package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"solution/internal/bucket"
	"solution/internal/httpserver"
	"syscall"
	"time"
)

type App struct {
	bucket *bucket.Client
	server *httpserver.MyServer
	config *Config
}

func main() {
	exitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app := new(App)
	app.config = NewConfig()
	app.bucket = bucket.NewClient(app.config.BucketServiceAddr, app.config.AccessKey,
		app.config.SecretKey, app.config.RefreshPeriodMin)
	err := app.bucket.SetBucket(bucket.BucketForData)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	objects, _ := app.bucket.GetAllObjects()
	users, _ := app.bucket.AggregateUsers(objects)

	err = app.bucket.StoreUsers(users, bucket.LocalPathToResult)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = app.bucket.SetBucketListener(exitCtx, bucket.BucketForData, users)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	userHandler := httpserver.NewUserHandler(users, app.bucket)
	router := httpserver.NewRouter(userHandler)
	app.server = httpserver.NewServer(app.config.HTTPPort, router)
	defer func() {
		ctx, cancel := context.WithTimeout(exitCtx, 5*time.Second)
		defer cancel()
		app.server.Stop(ctx)
	}()

	go app.server.Start()

	<-exitCtx.Done()
}
