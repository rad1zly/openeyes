// services/search_service.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
	"io/ioutil"
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
    searchType := s.determineQueryType(query)

    // Cari di ELK dulu
    elkResults, _ := s.searchElk(query, searchType)

    if searchType == "name" {
        // Jika dapat hasil dari ELK, masukkan ke response
        if len(elkResults) > 0 {
            for _, result := range elkResults {
                if result.Source == "leakosint" {
                    response.LeakosintResults = append(response.LeakosintResults, result)
                } else if result.Source == "linkedin" {
                    response.LinkedinResults = append(response.LinkedinResults, result)
                }
            }
        }

        // Tetap cari di API
        leakResults, _ := s.searchLeakosint(query)
        if len(leakResults) > 0 {
            response.LeakosintResults = append(response.LeakosintResults, leakResults...)
            // Save hasil baru ke ELK
            for _, result := range leakResults {
                s.saveToElk(result, "name")
            }
        }

        linkResults, _ := s.searchLinkedin(query)
        if len(linkResults) > 0 {
            response.LinkedinResults = append(response.LinkedinResults, linkResults...)
            // Save hasil baru ke ELK
            for _, result := range linkResults {
                s.saveToElk(result, "name")
            }
        }

    } else if searchType == "phone" {
        // Jika dapat hasil dari ELK, masukkan ke response
        if len(elkResults) > 0 {
            for _, result := range elkResults {
                if result.Source == "leakosint" {
                    response.LeakosintResults = append(response.LeakosintResults, result)
                } else if result.Source == "truecaller" {
                    response.TruecallerResults = append(response.TruecallerResults, result)
                }
            }
        }

        // Tetap cari di API
        leakResults, _ := s.searchLeakosint(query)
        if len(leakResults) > 0 {
            response.LeakosintResults = append(response.LeakosintResults, leakResults...)
            // Save hasil baru ke ELK
            for _, result := range leakResults {
                s.saveToElk(result, "phone")
            }
        }

        trueResults, _ := s.searchTruecaller(query)
        if len(trueResults) > 0 {
            response.TruecallerResults = append(response.TruecallerResults, trueResults...)
            // Save hasil baru ke ELK
            for _, result := range trueResults {
                s.saveToElk(result, "phone")
            }
        }

    } else if searchType == "nik" {
        // Untuk NIK, tetap gunakan flow yang lama
        if len(elkResults) > 0 {
            response.LeakosintResults = elkResults
            return response, nil
        }

        leakResults, _ := s.searchLeakosint(query)
        if len(leakResults) > 0 {
            response.LeakosintResults = leakResults
            // Save ke ELK
            for _, result := range leakResults {
                s.saveToElk(result, "nik")
            }
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
        for sourceName, sourceData := range list {
            if data, ok := sourceData.(map[string]interface{}); ok {
                result := models.SearchResult{
                    ID:        fmt.Sprintf("leakosint_%d", time.Now().UnixNano()),
                    Source:    "leakosint",  // Set static source untuk ELK
                    Data: map[string]interface{}{
                        "SourceName":   sourceName,  // Simpan nama source asli sebagai metadata
                        "Data":         data["Data"],
                        "InfoLeak":     data["InfoLeak"],
                        "NumOfResults": data["NumOfResults"],
                    },
                    Timestamp: time.Now(),
                }
                results = append(results, result)
            }
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
    reqBody := map[string]interface{}{
        "name":         query,
        "company_name": "",
        "job_title":    "",
        "location":     "",
        "keywords":     "",
        "limit":        15,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    fmt.Printf("LinkedIn Request: %s\n", string(jsonData))

    req, err := http.NewRequest("POST", s.config.LinkedinURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rapidapi-key", s.config.LinkedinAPIKey)
    req.Header.Set("x-rapidapi-host", s.config.LinkedinAPIHost)

    client := &http.Client{Timeout: time.Second * 10}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    fmt.Printf("LinkedIn Response: %s\n", string(body))

    // Parse response sebagai map dulu
    var responseData map[string]interface{}
    if err := json.Unmarshal(body, &responseData); err != nil {
        fmt.Printf("LinkedIn Parse Error: %v\n", err)
        return nil, err
    }

    // Convert ke SearchResult
    result := models.SearchResult{
        ID:        fmt.Sprintf("linkedin_%d", time.Now().UnixNano()),
        Source:    "linkedin",
        Data:      responseData,
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
    fmt.Printf("\nSaving to ELK - Original Source: %s\n", result.Source)

    // Periksa data sebelum disimpan
    jsonDataDebug, _ := json.MarshalIndent(result, "", "  ")
    fmt.Printf("Data to save: %s\n", string(jsonDataDebug))

    var indexName string
    switch result.Source {
    case "leakosint":
        indexName = "leakosint_data"
        fmt.Printf("Using leakosint index\n")
    case "linkedin":
        indexName = "linkedin_data"
        fmt.Printf("Using linkedin index\n")
    case "truecaller":
        indexName = "truecaller_data"
        fmt.Printf("Using truecaller index\n")
    default:
        fmt.Printf("Unknown source: %s\n", result.Source)
        return fmt.Errorf("unknown source: %s", result.Source)
    }

    documentData := map[string]interface{}{
        "id":        fmt.Sprintf("%s_%d", result.Source, time.Now().UnixNano()),
        "source":    result.Source, // Source asli dari API
        "data":      result.Data,
        "timestamp": time.Now(),
    }

    jsonData, _ := json.Marshal(documentData)
    fmt.Printf("Saving data: %s\n", string(jsonData))

    url := fmt.Sprintf("%s/%s/_doc", s.config.ElasticsearchURL, indexName)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.SetBasicAuth(s.config.ElasticsearchUser, s.config.ElasticsearchPassword)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

// Export fungsi dengan huruf kapital di awal
func (s *SearchService) TestElkConnection() error {
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

func (s *SearchService) searchElk(query string, searchType string) ([]models.SearchResult, error) {
    // Tentukan index yang akan dicari berdasarkan tipe pencarian
    var indexes []string
    switch searchType {
    case "name":
        indexes = []string{"leakosint_data", "linkedin_data"}
    case "phone":
        indexes = []string{"leakosint_data", "truecaller_data"}
    case "nik":
        indexes = []string{"leakosint_data"}
    }

    var searchResults []models.SearchResult
    
    // Cari di setiap index yang sesuai
    for _, indexName := range indexes {
        // Check if index exists
        reqCheck, _ := http.NewRequest("HEAD", fmt.Sprintf("%s/%s", s.config.ElasticsearchURL, indexName), nil)
        reqCheck.SetBasicAuth(s.config.ElasticsearchUser, s.config.ElasticsearchPassword)
        
        clientCheck := &http.Client{}
        respCheck, err := clientCheck.Do(reqCheck)
        if err != nil || respCheck.StatusCode == 404 {
            fmt.Printf("Index %s tidak ditemukan\n", indexName)
            continue
        }

        searchQuery := map[string]interface{}{
            "query": map[string]interface{}{
                "bool": map[string]interface{}{
                    "should": []map[string]interface{}{
                        {
                            "match": map[string]interface{}{
                                "data.Data.FullName": query,
                            },
                        },
                        {
                            "match": map[string]interface{}{
                                "data.full_name": query,
                            },
                        },
                        {
                            "match": map[string]interface{}{
                                "data.Data.Phone": query,
                            },
                        },
                        {
                            "match": map[string]interface{}{
                                "data.phoneInfo.e164Format": query,
                            },
                        },
                        {
                            "match": map[string]interface{}{
                                "data.Data.Passport": query,
                            },
                        },
                    },
                },
            },
        }

        url := fmt.Sprintf("%s/%s/_search", s.config.ElasticsearchURL, indexName)
        fmt.Printf("Mencari di index: %s\n", url)

        jsonData, _ := json.Marshal(searchQuery)
        req, _ := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        req.SetBasicAuth(s.config.ElasticsearchUser, s.config.ElasticsearchPassword)

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            continue
        }
        defer resp.Body.Close()

        body, _ := ioutil.ReadAll(resp.Body)
        fmt.Printf("Data di ELK: %s\n", string(body))

        var result map[string]interface{}
        json.Unmarshal(body, &result)

        if hits, ok := result["hits"].(map[string]interface{}); ok {
            if hitsList, ok := hits["hits"].([]interface{}); ok {
                for _, hit := range hitsList {
                    if hitMap, ok := hit.(map[string]interface{}); ok {
                        if source, ok := hitMap["_source"].(map[string]interface{}); ok {
                            searchResults = append(searchResults, models.SearchResult{
                                ID:        source["id"].(string),
                                Source:    source["source"].(string),
                                Data:      source["data"],
                                Timestamp: time.Now(),
                            })
                        }
                    }
                }
            }
        }
    }

    return searchResults, nil
}