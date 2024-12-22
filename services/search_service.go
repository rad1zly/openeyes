package services

import (
	"bytes"
	"context"
	"encoding/json"
	"openeyes/config"
	"openeyes/models"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type SearchService struct {
	es  *elasticsearch.Client
	cfg *config.Config
}

func NewSearchService(es *elasticsearch.Client, cfg *config.Config) *SearchService {
	return &SearchService{
		es:  es,
		cfg: cfg,
	}
}

func (s *SearchService) Search(query string) (*models.SearchResponse, error) {
	response := &models.SearchResponse{Query: query}
	queryType := s.determineQueryType(query)

	// Search in Elasticsearch first
	elkResults, err := s.searchElk(query)
	if err != nil {
		return nil, err
	}

	response.ElkResults.Results = elkResults
	response.ElkResults.Total = len(elkResults)

	// If no results in Elasticsearch, search in external APIs
	if len(elkResults) == 0 {
		switch queryType {
		case "name":
			leakosintResults, _ := s.searchLeakosint(query)
			linkedinResults, _ := s.searchLinkedin(query)

			response.LeakosintResults.Results = leakosintResults
			response.LeakosintResults.Total = len(leakosintResults)

			response.LinkedinResults.Results = linkedinResults
			response.LinkedinResults.Total = len(linkedinResults)

		case "nik":
			leakosintResults, _ := s.searchLeakosint(query)
			response.LeakosintResults.Results = leakosintResults
			response.LeakosintResults.Total = len(leakosintResults)

		case "phone":
			leakosintResults, _ := s.searchLeakosint(query)
			truecallerResults, _ := s.searchTruecaller(query)

			response.LeakosintResults.Results = leakosintResults
			response.LeakosintResults.Total = len(leakosintResults)

			response.TruecallerResults.Results = truecallerResults
			response.TruecallerResults.Total = len(truecallerResults)
		}
	}

	return response, nil
}

func (s *SearchService) determineQueryType(query string) string {
	if isNIK(query) {
		return "nik"
	}
	if isPhone(query) {
		return "phone"
	}
	return "name"
}

func (s *SearchService) searchElk(query string) ([]models.SearchResult, error) {
	ctx := context.Background()

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title", "content", "name", "phone", "nik"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := s.es.Search(
		s.es.Search.WithContext(ctx),
		s.es.Search.WithIndex("search_index"),
		s.es.Search.WithBody(&buf),
		s.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	var searchResults []models.SearchResult
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		searchResults = append(searchResults, models.SearchResult{
			ID:        hitMap["_id"].(string),
			Source:    "elk",
			Data:      source,
			Timestamp: time.Now(),
		})
	}

	return searchResults, nil
}

func (s *SearchService) searchLeakosint(query string) ([]models.SearchResult, error) {
	// Implement Leakosint API integration
	// This is a placeholder implementation
	return []models.SearchResult{}, nil
}

func (s *SearchService) searchLinkedin(query string) ([]models.SearchResult, error) {
	// Implement LinkedIn API integration
	// This is a placeholder implementation
	return []models.SearchResult{}, nil
}

func (s *SearchService) searchTruecaller(query string) ([]models.SearchResult, error) {
	// Implement Truecaller API integration
	// This is a placeholder implementation
	return []models.SearchResult{}, nil
}

func isNIK(query string) bool {
	// Implement NIK validation
	// Example: check if string is 16 digits
	return len(query) == 16
}

func isPhone(query string) bool {
	// Implement phone number validation
	// Example: check if string starts with country code
	return len(query) >= 10 && query[0] == '6'
}
