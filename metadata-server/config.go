package main

import (
	"os"
	"strconv"
)

var (
	PORT            string
	UPLOADDIR       string
	MAXNAMESIZE   	int64
	DB_PATH 		string
	ISDEV			bool
)

func init() {
	PORT = getEnv("PORT", "7676")
	UPLOADDIR = getEnv("UPLOAD_DIR", "./uploads")
	
	//TODO: check this
	maxSizeStr := getEnv("MAX_UPLOAD_SIZE", "1000")
	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		maxSize = 1000
	}
	MAXNAMESIZE = maxSize
	DB_PATH = "db"
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

