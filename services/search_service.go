// services/search_service.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "openeyes/models"
	"crypto/sha256"
	"encoding/hex"
	"openeyes/config"
	"github.com/elastic/go-elasticsearch/v8"
)

type SearchService struct {
    config *config.Config
}

func NewSearchService(cfg *config.Config) *SearchService {
    return &SearchService{
        config: cfg,
    }
}

// Fungsi utama Search
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

// Implementasi searchLeakosint
func (s *SearchService) searchLeakosint(query string) ([]models.SearchResult, error) {
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

    req, err := http.NewRequest("POST", "https://leakosintapi.com/", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

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

    // Save to ELK
    if err := s.saveToElk(results, s.determineQueryType(query)); err != nil {
		fmt.Printf("Error saving to ELK: %v\n", err)
	}

    return results, nil
}

// Helper functions
func (s *SearchService) determineQueryType(query string) models.QueryType {
    if isNIK(query) {
        return models.QueryTypeNIK
    }
    if isPhone(query) {
        return models.QueryTypePhone
    }
    return models.QueryTypeName
}

func isNIK(query string) bool {
    return len(query) == 16
}

func isPhone(query string) bool {
    return len(query) >= 10 && query[0] == '6'
}

// Placeholder functions untuk API lain (akan diimplementasi nanti)
func (s *SearchService) searchLinkedin(query string) ([]models.SearchResult, error) {
    reqBody := models.LinkedInRequest{
        Name:        query,
        CompanyName: "",
        JobTitle:    "",
        Location:    "",
        Keywords:    "",
        Limit:       5,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", 
        "https://fresh-linkedin-profile-data.p.rapidapi.com/google-full-profiles", 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Add("x-rapidapi-key", s.config.LinkedinAPIKey)
    req.Header.Add("x-rapidapi-host", "fresh-linkedin-profile-data.p.rapidapi.com")
    req.Header.Add("Content-Type", "application/json")

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
        result := models.SearchResult{
            Source:    "linkedin",
            Data:     profile,
            Timestamp: time.Now(),
        }
        results = append(results, result)

        // Save to ELK
        if err := s.saveToElk([]models.SearchResult{result}, models.QueryTypeName); err != nil {
			fmt.Printf("Error saving to ELK: %v\n", err)
		}
    }

    return results, nil
}

func (s *SearchService) searchTruecaller(query string) ([]models.SearchResult, error) {
    url := fmt.Sprintf("https://truecaller-data2.p.rapidapi.com/search/%s", query)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    // Add headers
    req.Header.Add("x-rapidapi-key", s.config.TruecallerAPIKey)
    req.Header.Add("x-rapidapi-host", "truecaller-data2.p.rapidapi.com")

    client := &http.Client{Timeout: time.Second * 10}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("truecaller API returned status: %d", resp.StatusCode)
    }

    var truecallerResp models.TruecallerResponse
    if err := json.NewDecoder(resp.Body).Decode(&truecallerResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %v", err)
    }

    // Convert to SearchResult format
    result := models.SearchResult{
        Source:    "truecaller",
        Data:     truecallerResp.Data,
        Timestamp: time.Now(),
    }

    // Save to ELK
    if err := s.saveToElk([]models.SearchResult{result}, models.QueryTypePhone); err != nil {
		fmt.Printf("Error saving to ELK: %v\n", err)
	}

    return []models.SearchResult{result}, nil
}

func (s *SearchService) saveToElk(results []models.SearchResult, queryType models.QueryType) error {
    // Buat basic auth transport
    tr := &http.Transport{}
    client := &http.Client{Transport: tr}

    // Siapkan bulk request
    var bulkRequestBody strings.Builder
    for _, result := range results {
        // Action line
        action := map[string]interface{}{
            "index": map[string]interface{}{
                "_index": fmt.Sprintf("%s_data", queryType),
                "_id":    generateID(result),
            },
        }
        actionLine, _ := json.Marshal(action)
        bulkRequestBody.WriteString(string(actionLine) + "\n")

        // Document line
        docLine, _ := json.Marshal(result)
        bulkRequestBody.WriteString(string(docLine) + "\n")
    }

    // Buat request
    req, err := http.NewRequest("POST", 
        fmt.Sprintf("%s/_bulk", s.config.ElasticsearchURL), 
        strings.NewReader(bulkRequestBody.String()))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        return err
    }

    // Tambahkan headers
    req.Header.Set("Content-Type", "application/x-ndjson")
    req.SetBasicAuth(s.config.ElasticsearchUser, s.config.ElasticsearchPassword)

    // Kirim request
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending request: %v", err)
        return err
    }
    defer resp.Body.Close()

    // Baca response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response: %v", err)
        return err
    }

    // Log response untuk debugging
    log.Printf("Elasticsearch response: %s", string(body))

    // Check status code
    if resp.StatusCode >= 400 {
        return fmt.Errorf("elasticsearch returned status: %d, body: %s", 
            resp.StatusCode, string(body))
    }

    log.Printf("Successfully saved %d documents to Elasticsearch", len(results))
    return nil
}

// Helper function untuk generate ID
func generateID(result models.SearchResult) string {
    // Bisa menggunakan hash dari kombinasi data untuk memastikan uniqueness
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%s-%v-%s", result.Source, result.Data, result.Timestamp.String())))
    return hex.EncodeToString(h.Sum(nil))
}