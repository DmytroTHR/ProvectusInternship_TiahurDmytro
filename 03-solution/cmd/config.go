package main

import "os"

type Config struct {
	AccessKey, SecretKey string
}

func NewConfig() *Config {
	return &Config{
		AccessKey: os.Getenv("MINIO_ROOT_USER"),
		SecretKey: os.Getenv("MINIO_ROOT_PASSWORD"),
	}
}
