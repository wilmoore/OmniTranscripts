package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	APIKey           string
	AssemblyAIAPIKey string
	WhisperServerURL string
	WhisperModelPath string
	WorkDir          string
	MaxVideoLength   int
	FreeJobLimit     int
}

func Load() *Config {
	godotenv.Load()

	maxLength, _ := strconv.Atoi(getEnv("MAX_VIDEO_LENGTH", "1800"))
	freeLimit, _ := strconv.Atoi(getEnv("FREE_JOB_LIMIT", "5"))

	return &Config{
		Port:             getEnv("PORT", "3000"),
		APIKey:           getEnv("API_KEY", "your-api-key-here"),
		AssemblyAIAPIKey: getEnv("ASSEMBLYAI_API_KEY", ""),
		WhisperServerURL: getEnv("WHISPER_SERVER_URL", ""),
		WhisperModelPath: getEnv("WHISPER_MODEL_PATH", ""),
		WorkDir:          getEnv("WORK_DIR", "/tmp/omnitranscripts"),
		MaxVideoLength:   maxLength,
		FreeJobLimit:     freeLimit,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
