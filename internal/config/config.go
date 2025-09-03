package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration loaded from env/flags.
type Config struct {
	Provider                string
	LiteLLMBaseURL          string
	LiteLLMAPIKey           string
	Model                   string
	Temperature             float64
	TopP                    float64
	DBPath                  string
	SystemPrompt            string
	ModelContextTokens      int
	EnableProviderWebsearch bool
	DropSamplingParams      bool
	DebugPrompts            bool
	AllowLocalShell         bool
}

// Load returns configuration with env values and sane defaults.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Provider:                getenvDefault("LLM_PROVIDER", "litellm"),
		LiteLLMBaseURL:          getenvDefault("LITELLM_BASE_URL", "http://localhost:4000"),
		LiteLLMAPIKey:           os.Getenv("LITELLM_API_KEY"),
		Model:                   os.Getenv("LLM_MODEL"),
		DBPath:                  getenvDefault("DB_PATH", "clichat.db"),
		SystemPrompt:            getenvDefault("SYSTEM_PROMPT", "You are a concise, helpful CLI assistant."),
		EnableProviderWebsearch: getBool("ENABLE_PROVIDER_WEBSEARCH", false),
		DropSamplingParams:      getBool("DROP_SAMPLING_PARAMS", false),
		DebugPrompts:            getBool("DEBUG_PROMPTS", false),
		AllowLocalShell:         getBool("ALLOW_LOCAL_SHELL", false),
	}

	cfg.Temperature = getFloat("TEMPERATURE", 0.2)
	cfg.TopP = getFloat("TOP_P", 1.0)
	cfg.ModelContextTokens = getInt("MODEL_CONTEXT_TOKENS", 0)

	return cfg, nil
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
