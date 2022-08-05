package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AccessKey, SecretKey, HTTPPort, BucketServiceAddr string
	RefreshPeriodMin                                  time.Duration
}

func NewConfig() *Config {
	refreshPeriod, err := strconv.ParseUint(os.Getenv("REFRESH_PERIOD_MIN"), 10, 16)
	if err != nil {
		refreshPeriod = 3
	}
	return &Config{
		AccessKey:         os.Getenv("MINIO_ROOT_USER"),
		SecretKey:         os.Getenv("MINIO_ROOT_PASSWORD"),
		HTTPPort:          os.Getenv("HTTP_SERVER_PORT"),
		BucketServiceAddr: os.Getenv("BUCKET_SERVICE_ADDR"),
		RefreshPeriodMin:  time.Duration(refreshPeriod) * time.Minute,
	}
}
