package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	PORT          string
	UPLOADDIR     string
	MAXNAMESIZE   int64
	DB_PATH       string
	ISDEV         bool
	GLOBAL_CONFIG *Config
)

type Config struct {
	Server ServerConfig
	RustFS RustFSConfig
}

type ServerConfig struct {
	Port string
}

type RustFSConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	// Region          string
	PresignDuration time.Duration
	UsePathStyle    bool // must be true for RustFS / MinIO
}

func init() {
	PORT = getEnv("PORT", "7678")
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

	// TODO: consolidate everything in Config
	//
	presignMins, err := strconv.Atoi(getEnv("RUSTFS_PRESIGN_MINUTES", "15"))
	if err != nil {
		log.Fatalf("invalid RUSTFS_PRESIGN_MINUTES: %v", err)
	}

	GLOBAL_CONFIG = &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		//checking if env vars are set in RustFSConfig constructor
		RustFS: RustFSConfig{
			Endpoint:  getEnv("RUSTFS_ENDPOINT", "http://127.0.0.1:9000"),
			AccessKey: getEnv("RUSTFS_ACCESS_KEY", "rustfsadmin"),
			SecretKey: getEnv("RUSTFS_SECRET_KEY", "rustfsadmin"),
			Bucket:    getEnv("RUSTFS_BUCKET", "tests"),
			// Region:          getEnv("RUSTFS_REGION", "us-east-1"),
			PresignDuration: time.Duration(presignMins) * time.Minute,
			UsePathStyle:    true,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return v
}
