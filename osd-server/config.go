package main

import (
	"fmt"
	"os"
	"strconv"
)

var (
	PORT              string
	META_PORT         string
	UPLOADDIR         string
	UPLOADPVW         string
	MAXUPLOADSIZE     int64
	METADATASERVERURL string
	ISDEV             bool
)

const (
	PREVIEW_MAX_WIDTH  = 150 
	PREVIEW_MAX_HEIGHT = 200
)

func init() {
	PORT = getEnv("PORT", "8280")
	META_PORT = getEnv("PORT", "7676")
	UPLOADDIR = getEnv("UPLOAD_DIR", "./store")

	maxSizeStr := getEnv("MAX_UPLOAD_SIZE", "10485760")
	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		maxSize = 10 << 20
	}
	MAXUPLOADSIZE = maxSize

	METADATASERVERURL = getEnv("METADATA_SERVER_URL", fmt.Sprintf("http://metadata-server:%s", META_PORT))
	if getEnv("ISDEV", "true") == "true" {
		ISDEV = true
	} else {
		ISDEV = false
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
