package models

import "time"

type SearchResult struct {
	ID        string      `json:"id"`
	Source    string      `json:"source"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type SearchResponse struct {
	Query      string `json:"query"`
	ElkResults struct {
		Total   int            `json:"total"`
		Results []SearchResult `json:"results"`
	} `json:"elk_results"`
	LeakosintResults struct {
		Total   int            `json:"total"`
		Results []SearchResult `json:"results"`
	} `json:"leakosint_results"`
	LinkedinResults struct {
		Total   int            `json:"total"`
		Results []SearchResult `json:"results"`
	} `json:"linkedin_results,omitempty"`
	TruecallerResults struct {
		Total   int            `json:"total"`
		Results []SearchResult `json:"results"`
	} `json:"truecaller_results,omitempty"`
}

type QueryType string

const (
    QueryTypeName  QueryType = "name"
    QueryTypeNIK   QueryType = "nik"
    QueryTypePhone QueryType = "phone"
)