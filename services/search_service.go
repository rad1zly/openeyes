// services/search_service.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "openeyes/config"
    "openeyes/models"
)

type SearchService struct {
    config *config.Config
}

func NewSearchService(cfg *config.Config) *SearchService {
    return &SearchService{
        config: cfg,
    }
}

func (s *SearchService) Search(query string) (*models.SearchResponse, error) {
    queryType := s.determineQueryType(query)
    response := &models.SearchResponse{Query: query}

    switch queryType {
    case "name":
        leakosintResults, _ := s.searchLeakosint(query)
        linkedinResults, _ := s.searchLinkedin(query)
        
        response.LeakosintResults.Results = leakosintResults
        response.LinkedinResults.Results = linkedinResults

    case "nik":
        leakosintResults, _ := s.searchLeakosint(query)
        response.LeakosintResults.Results = leakosintResults

    case "phone":
        leakosintResults, _ := s.searchLeakosint(query)
        truecallerResults, _ := s.searchTruecaller(query)
        
        response.LeakosintResults.Results = leakosintResults
        response.TruecallerResults.Results = truecallerResults
    }

    return response, nil
}

func (s *SearchService) searchLeakosint(query string) ([]models.SearchResult, error) {
    reqBody := models.LeakosintRequest{
        Token:   s.config.LeakosintAPIKey,
        Request: query,
        Limit:   100,
        Lang:    "en",
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", s.config.LeakosintURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: time.Second * 10}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var leakosintResp models.LeakosintResponse
    if err := json.NewDecoder(resp.Body).Decode(&leakosintResp); err != nil {
        return nil, err
    }

    var results []models.SearchResult
    for source, sourceData := range leakosintResp.List {
        for _, data := range sourceData.Data {
            results = append(results, models.SearchResult{
                Source:    source,
                Data:     data,
                Timestamp: time.Now(),
            })
        }
    }

    return results, nil
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

func isNIK(query string) bool {
    return len(query) == 16
}

func isPhone(query string) bool {
    return len(query) >= 10 && query[0] == '6'
}