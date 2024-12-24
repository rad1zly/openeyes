// services/search_service.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
	//"io/ioutil"
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
 
    // Untuk pencarian nama
    if !isNIK(query) && !isPhone(query) {
        // Cari di Leakosint
        leakResults, _ := s.searchLeakosint(query)
        if len(leakResults) > 0 {
            for _, result := range leakResults {
                // Simpan hasil ke ELK
                s.saveToElk(result, "name")
            }
            response.LeakosintResults = leakResults
        }
 
        // Cari di LinkedIn
        linkResults, _ := s.searchLinkedin(query)
        if len(linkResults) > 0 {
            for _, result := range linkResults {
                // Simpan hasil ke ELK 
                s.saveToElk(result, "name")
            }
            response.LinkedinResults = linkResults
        }
 
    // Untuk pencarian NIK
    } else if isNIK(query) {
        // Cari di Leakosint
        leakResults, _ := s.searchLeakosint(query)
        if len(leakResults) > 0 {
            for _, result := range leakResults {
                // Simpan hasil ke ELK
                s.saveToElk(result, "nik")
            }
            response.LeakosintResults = leakResults
        }
 
    // Untuk pencarian nomor telepon
    } else if isPhone(query) {
        // Cari di Leakosint
        leakResults, _ := s.searchLeakosint(query)
        if len(leakResults) > 0 {
            for _, result := range leakResults {
                // Simpan hasil ke ELK
                s.saveToElk(result, "phone")
            }
            response.LeakosintResults = leakResults
        }
 
        // Cari di Truecaller
        trueResults, _ := s.searchTruecaller(query)
        if len(trueResults) > 0 {
            for _, result := range trueResults {
                // Simpan hasil ke ELK
                s.saveToElk(result, "phone")
            }
            response.TruecallerResults = trueResults
        }
    }
 
    return response, nil
 }

func (s *SearchService) searchLeakosint(query string) ([]models.SearchResult, error) {
    apiResponse, err := s.queryLeakosintAPI(query)
    if err != nil {
        return nil, err
    }

    var results []models.SearchResult
    if list, ok := apiResponse["List"].(map[string]interface{}); ok {
        for source, sourceData := range list {
            result := models.SearchResult{
                ID:        fmt.Sprintf("%s_%d", source, time.Now().UnixNano()),
                Source:    source,
                Data:      sourceData,
                Timestamp: time.Now(),
            }
            results = append(results, result)
        }
    }

    return results, nil
}

func (s *SearchService) queryLeakosintAPI(query string) (models.LeakosintResponse, error) {
    reqBody := models.LeakosintRequest{
        Token:   s.config.LeakosintAPIKey,
        Request: query,
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

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    var apiResponse models.LeakosintResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return nil, fmt.Errorf("error decoding response: %v", err)
    }

    return apiResponse, nil
}

func (s *SearchService) searchLinkedin(query string) ([]models.SearchResult, error) {
    data := map[string]interface{}{
        "name": query,
        "company_name": "",
        "job_title": "",
        "location": "",
        "keywords": "",
        "limit": 15,
    }
 
    jsonData, _ := json.Marshal(data)
    fmt.Printf("LinkedIn Request: %s\n", string(jsonData))
 
    req, _ := http.NewRequest("POST", s.config.LinkedinURL, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rapidapi-key", s.config.LinkedinAPIKey)
    req.Header.Set("x-rapidapi-host", s.config.LinkedinAPIHost)
 
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("LinkedIn Error: %v\n", err)
        return nil, err
    }
    defer resp.Body.Close()
 
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Printf("LinkedIn Response: %s\n", string(body))
 
    var apiResponse []interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        fmt.Printf("LinkedIn Parse Error: %v\n", err)
        return nil, err
    }
 
    var results []models.SearchResult
    for _, data := range apiResponse {
        results = append(results, models.SearchResult{
            ID:        fmt.Sprintf("linkedin_%d", time.Now().UnixNano()),
            Source:    "linkedin",
            Data:      data,
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

func (s *SearchService) saveToElk(result models.SearchResult, sourceType string) error {
    // Siapkan data yang akan disimpan
    jsonData, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("error marshaling data: %v", err)
    }

    // Buat request ke Elasticsearch
    url := fmt.Sprintf("%s/%s_data/_doc", s.config.ElasticsearchURL, sourceType)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }

    // Set header dan auth
    req.Header.Set("Content-Type", "application/json")
    req.SetBasicAuth(s.config.ElasticsearchUser, s.config.ElasticsearchPassword)

    // Kirim request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error saving to elasticsearch: %v", err)
    }
    defer resp.Body.Close()

    // Cek response
    if resp.StatusCode >= 400 {
        return fmt.Errorf("elasticsearch error: status code %d", resp.StatusCode)
    }

    return nil
}

func (s *SearchService) testElkConnection() error {
    req, err := http.NewRequest("GET", s.config.ElasticsearchURL, nil)
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }

    req.SetBasicAuth(s.config.ElasticsearchUser, s.config.ElasticsearchPassword)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error connecting to elasticsearch: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("elasticsearch returned status: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return fmt.Errorf("error decoding response: %v", err)
    }

    fmt.Printf("Connected to Elasticsearch version: %v\n", result["version"].(map[string]interface{})["number"])
    return nil
}