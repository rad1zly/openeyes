// models/leakosint.go
package models

type LeakosintRequest struct {
    Token   string `json:"token"`
    Request string `json:"request"`
    Limit   int    `json:"limit"`
    Lang    string `json:"lang"`
}

type LeakosintSourceData struct {
    Data         []interface{} `json:"Data"`         // Menggunakan interface{} agar fleksibel
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