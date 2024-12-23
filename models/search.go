package models

import "time"

type SearchResult struct {
    ID        string      `json:"id"`
    Source    string      `json:"source"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
}

type SearchResponse struct {
    Query             string         `json:"query"`
    LeakosintResults  []SearchResult `json:"leakosint_results"`
    LinkedinResults   []SearchResult `json:"linkedin_results,omitempty"`
    TruecallerResults []SearchResult `json:"truecaller_results,omitempty"`
}

type QueryType string

const (
    QueryTypeName  QueryType = "name"
    QueryTypeNIK   QueryType = "nik"
    QueryTypePhone QueryType = "phone"
)

// Model untuk request ke Leakosint
type LeakosintRequest struct {
    Token   string `json:"token"`
    Request string `json:"request"`
    Limit   int    `json:"limit"`
    Lang    string `json:"lang"`
}

type LeakosintSourceData struct {
    Data         []interface{} `json:"Data"`
    NumOfResults int          `json:"NumOfResults"`
    InfoLeak     string       `json:"InfoLeak"`
}

type LeakosintResponse struct {
    NumOfResults     int                           `json:"NumOfResults"`
    List            map[string]LeakosintSourceData `json:"List"`
    NumOfDatabase   int                           `json:"NumOfDatabase"`
    SearchTime      float64                       `json:"search time"`
    Price           int                           `json:"price"`
    FreeRequestsLeft int                          `json:"free_requests_left"`
}

// Model untuk request ke LinkedIn
type LinkedInRequest struct {
    Name        string `json:"name"`
    CompanyName string `json:"company_name"`
    JobTitle    string `json:"job_title"`
    Location    string `json:"location"`
    Keywords    string `json:"keywords"`
    Limit       int    `json:"limit"`
}

type LinkedInProfile struct {
    MatchScore      int    `json:"_match_score"`
    FullName        string `json:"full_name"`
    FirstName       string `json:"first_name"`
    LastName        string `json:"last_name"`
    Headline        string `json:"headline"`
    Location        string `json:"location"`
    ProfileImageURL string `json:"profile_image_url"`
    // ... tambahkan field lain sesuai kebutuhan
}

// Model untuk response Truecaller
type TruecallerResponse struct {
    Data struct {
        AddressInfo struct {
            Address     string `json:"address"`
            Street      string `json:"street"`
            ZipCode     string `json:"zipCode"`
            CountryCode string `json:"countryCode"`
            City        string `json:"city"`
            TimeZone    string `json:"timeZone"`
        } `json:"addressInfo"`
        BasicInfo struct {
            About    string `json:"about"`
            Gender   string `json:"gender"`
            Image    string `json:"image"`
            JobTitle string `json:"jobTitle"`
            Name     struct {
                AltName  string `json:"altName"`
                FullName string `json:"fullName"`
            } `json:"name"`
        } `json:"basicInfo"`
        PhoneInfo struct {
            E164Format      string `json:"e164Format"`
            NumberType      string `json:"numberType"`
            NationalFormat  string `json:"nationalFormat"`
            DialingCode     int    `json:"dialingCode"`
            CountryCode     string `json:"countryCode"`
            Carrier        string `json:"carrier"`
            Type           string `json:"type"`
        } `json:"phoneInfo"`
    } `json:"data"`
    Error int `json:"error"`
}