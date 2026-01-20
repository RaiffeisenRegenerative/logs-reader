package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	SSHHost        string
	SSHUser        string
	SSHKey         string
	EmailPort      int
	EmailLogPath   string
	NextjsPort     int
	NextjsLogFiles []string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: No .env file found, falling back to system env")
	}

	getEnv := func(key string, fallback string) string {
		val := os.Getenv(key)
		if val == "" {
			return fallback
		}
		return val
	}

	parseInt := func(key string) int {
		valStr := os.Getenv(key)
		val, err := strconv.Atoi(valStr)
		if err != nil {
			fmt.Printf("Invalid integer for %s: %s\n", key, valStr)
			os.Exit(1)
		}
		return val
	}

	return &Config{
		SSHHost:        getEnv("SSH_HOST", "r1.peaknetworks.net"),
		SSHUser:        getEnv("SSH_USER", "root"),
		SSHKey:         getEnv("SSH_KEY", "~/.ssh/id_rsa"),
		EmailPort:      parseInt("EMAIL_LOG_PORT"),
		EmailLogPath:   getEnv("EMAIL_LOG_PATH", "email_logs/debug.log"),
		NextjsPort:     parseInt("NEXTJS_LOG_PORT"),
		NextjsLogFiles: strings.Split(getEnv("NEXTJS_LOG_FILES", ""), ","),
	}, nil
}
