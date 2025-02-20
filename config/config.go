// config/config.go
package config

import (
	"os"
	"log"
	"github.com/joho/godotenv"
)

type Config struct {
    // Elasticsearch Configuration
    ElasticsearchURL      string
    ElasticsearchUser     string
    ElasticsearchPassword string

    // Leakosint Configuration
    LeakosintURL         string
    LeakosintAPIKey      string

    // LinkedIn Configuration
    LinkedinURL          string
    LinkedinAPIKey       string
    LinkedinAPIHost      string

    // Truecaller Configuration
    TruecallerURL        string
    TruecallerAPIKey     string
    TruecallerAPIHost    string
}

func LoadConfig() *Config {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    return &Config{
        LeakosintURL:      "https://leakosintapi.com/",
        LeakosintAPIKey:   os.Getenv("LEAKOSINT_API_KEY"),
        LinkedinURL:       "https://fresh-linkedin-profile-data.p.rapidapi.com/google-full-profiles",
        LinkedinAPIKey:    os.Getenv("LINKEDIN_API_KEY"),
        LinkedinAPIHost:   "fresh-linkedin-profile-data.p.rapidapi.com",
		TruecallerURL: 	   "https://truecaller-data2.p.rapidapi.com/search",
		TruecallerAPIKey:  os.Getenv("TRUECALLER_API_KEY"),
		TruecallerAPIHost: "truecaller-data2.p.rapidapi.com",
        ElasticsearchURL:   "http://localhost:9200",
        ElasticsearchUser:     "elastic",
        ElasticsearchPassword: os.Getenv("ELASTICSEARCH_PASSWORD"),
	}
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}