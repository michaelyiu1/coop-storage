package main

import (
	"os"
	"strconv"
	"fmt"
)

var (
	PORT            string
	META_PORT		string
	UPLOADDIR       string
	MAXUPLOADSIZE   int64
	METADATASERVERURL string
	 ISDEV bool
)

func init() {
	PORT = getEnv("PORT", "8280")
	META_PORT = getEnv("META_PORT", "7676") //Should be META_PORT??
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

