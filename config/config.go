package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func GetConfigLevel() logrus.Level {
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.DebugLevel
	}
	return level
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	return port
}

func GetRegion() string {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-west-1"
	}
	return region
}

func GetRequestIdHeader() string {
	return os.Getenv("REQUEST_ID_HEADER")
}
