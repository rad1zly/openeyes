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
    response := &models.SearchResponse{Query: query}

    // Untuk kasus pencarian nama, cari di Leakosint dan LinkedIn
    if !isNIK(query) && !isPhone(query) {
        // Cari di Leakosint
        leakosintResults, err := s.searchLeakosint(query)
        if err == nil && len(leakosintResults) > 0 {
            response.LeakosintResults = leakosintResults
        }

        // Cari di LinkedIn
        linkedinResults, err := s.searchLinkedin(query)
        if err == nil && len(linkedinResults) > 0 {
            response.LinkedinResults = linkedinResults
        }
        return response, nil
    }

    // Untuk kasus pencarian NIK
    if isNIK(query) {
        leakosintResults, err := s.searchLeakosint(query)
        if err == nil && len(leakosintResults) > 0 {
            response.LeakosintResults = leakosintResults
        }
        return response, nil
    }

    // Untuk kasus pencarian nomor telepon
    if isPhone(query) {
        // Cari di Leakosint
        leakosintResults, err := s.searchLeakosint(query)
        if err == nil && len(leakosintResults) > 0 {
            response.LeakosintResults = leakosintResults
        }

        // Cari di Truecaller
        truecallerResults, err := s.searchTruecaller(query)
        if err == nil && len(truecallerResults) > 0 {
            response.TruecallerResults = truecallerResults
        }
        return response, nil
    }

    return response, nil
}

func (s *SearchService) searchLeakosint(query string) ([]models.SearchResult, error) {
    reqBody := models.LeakosintRequest{
        Token:   s.config.LeakosintAPIKey,  // dari .env
        Request: query,                    // dari parameter
        Limit:   100,
        Lang:    "en",
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", s.config.LeakosintURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    // Hanya perlu content-type json
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: time.Second * 10}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    var leakosintResp models.LeakosintResponse
    if err := json.NewDecoder(resp.Body).Decode(&leakosintResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %v", err)
    }

    var results []models.SearchResult
    if leakosintResp.List != nil {
        for source, sourceData := range leakosintResp.List {
            for _, data := range sourceData.Data {
                result := models.SearchResult{
                    Source:    source,
                    Data:     data,
                    Timestamp: time.Now(),
                }
                results = append(results, result)
            }
        }
    }

    return results, nil
}

func (s *SearchService) searchLinkedin(query string) ([]models.SearchResult, error) {
    reqBody := models.LinkedInRequest{
        Name:    query,
        Limit:   5,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", s.config.LinkedinURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rapidapi-key", s.config.LinkedinAPIKey)
    req.Header.Set("x-rapidapi-host", s.config.LinkedinAPIHost)

    client := &http.Client{Timeout: time.Second * 10}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    var profiles []models.LinkedInProfile
    if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
        return nil, fmt.Errorf("error decoding response: %v", err)
    }

    var results []models.SearchResult
    for _, profile := range profiles {
        results = append(results, models.SearchResult{
            Source:    "linkedin",
            Data:     profile,
            Timestamp: time.Now(),
        })
    }

    return results, nil
}

func (s *SearchService) searchTruecaller(query string) ([]models.SearchResult, error) {
    url := fmt.Sprintf("%s/%s", s.config.TruecallerURL, query)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Add("x-rapidapi-key", s.config.TruecallerAPIKey)
    req.Header.Add("x-rapidapi-host", s.config.TruecallerAPIHost)

    client := &http.Client{Timeout: time.Second * 10}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    var truecallerResp models.TruecallerResponse
    if err := json.NewDecoder(resp.Body).Decode(&truecallerResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %v", err)
    }

    result := models.SearchResult{
        Source:    "truecaller",
        Data:     truecallerResp.Data,
        Timestamp: time.Now(),
    }

    return []models.SearchResult{result}, nil
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