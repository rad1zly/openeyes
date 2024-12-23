// config/config.go
package config

import "os"

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
    return &Config{
        // Elasticsearch Config
        ElasticsearchURL:      getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
        ElasticsearchUser:     getEnv("ELASTICSEARCH_USER", "elastic"),
        ElasticsearchPassword: getEnv("ELASTICSEARCH_PASSWORD", ""),

        // Leakosint Config
        LeakosintURL:    getEnv("LEAKOSINT_URL", "https://leakosintapi.com/"),
        LeakosintAPIKey: getEnv("LEAKOSINT_API_KEY", ""),

        // LinkedIn Config
        LinkedinURL:     getEnv("LINKEDIN_URL", "https://fresh-linkedin-profile-data.p.rapidapi.com/google-full-profiles"),
        LinkedinAPIKey:  getEnv("LINKEDIN_API_KEY", ""),
        LinkedinAPIHost: getEnv("LINKEDIN_API_HOST", "fresh-linkedin-profile-data.p.rapidapi.com"),

        // Truecaller Config
        TruecallerURL:     getEnv("TRUECALLER_URL", "https://truecaller-data2.p.rapidapi.com/search"),
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