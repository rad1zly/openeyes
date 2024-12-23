package models

import "time"

type SearchResult struct {
	ID        string      `json:"id"`
	Source    string      `json:"source"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

package models

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