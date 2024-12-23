package config

import (
	"os"
)

type Config struct {
	ElasticsearchURL string
	LeakosintAPIKey  string
	LinkedinAPIKey   string
	TruecallerAPIKey string
    TruecallerAPIHost  string
}
}

func LoadConfig() *Config {
	return &Config{
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		LeakosintAPIKey:  getEnv("LEAKOSINT_API_KEY", ""),
		LinkedinAPIKey:   getEnv("LINKEDIN_API_KEY", ""),
		TruecallerAPIKey: getEnv("TRUECALLER_API_KEY", ""),
		TruecallerAPIKey:  getEnv("TRUECALLER_API_KEY", ""),
        TruecallerAPIHost: getEnv("TRUECALLER_API_HOST", "truecaller-data2.p.rapidapi.com"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
