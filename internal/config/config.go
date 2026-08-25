package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	ClamAVHost      string
	ClamAVPort      string
	ClamAVTimeout   time.Duration
	MaxFileSizeBytes int64
}

func Load() Config {
	return Config{
		Port:            envOrDefault("PORT", "8080"),
		ClamAVHost:      envOrDefault("CLAMAV_HOST", "clamav"),
		ClamAVPort:      envOrDefault("CLAMAV_PORT", "3310"),
		ClamAVTimeout:   parseDuration(envOrDefault("CLAMAV_TIMEOUT_SECONDS", "30")),
		MaxFileSizeBytes: parseInt64(envOrDefault("MAX_FILE_SIZE_BYTES", "26214400")),
	}
}

func (c Config) ClamAVAddr() string {
	return c.ClamAVHost + ":" + c.ClamAVPort
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(seconds string) time.Duration {
	s, err := strconv.Atoi(seconds)
	if err != nil {
		return 30 * time.Second
	}
	return time.Duration(s) * time.Second
}

func parseInt64(val string) int64 {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 26214400
	}
	return n
}
