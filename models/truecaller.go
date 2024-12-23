// models/truecaller.go
package models

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
            Name struct {
                AltName  string `json:"altName"`
                FullName string `json:"fullName"`
            } `json:"name"`
        } `json:"basicInfo"`
        PhoneInfo struct {
            E164Format      string `json:"e164Format"`
            NumberType     string `json:"numberType"`
            NationalFormat string `json:"nationalFormat"`
            DialingCode    int    `json:"dialingCode"`
            CountryCode    string `json:"countryCode"`
            SpamScore      int    `json:"spamScore"`
            SpamType       string `json:"spamType"`
            Carrier        string `json:"carrier"`
            Type          string `json:"type"`
        } `json:"phoneInfo"`
        Score          float64 `json:"score"`
        SpamInfo struct {
            SpamScore   int      `json:"spamScore"`
            SpamType    string   `json:"spamType"`
            SpamStats struct {
                NumReports          int     `json:"numReports"`
                NumReports60days    int     `json:"numReports60days"`
                NumSearches60days   int     `json:"numSearches60days"`
                NumCallsHourly      *int    `json:"numCallsHourly"`
                NumCalls60days      int     `json:"numCalls60days"`
                NumCallsNotAnswered int     `json:"numCallsNotAnswered"`
                NumCallsAnswered    int     `json:"numCallsAnswered"`
                NumMessages60days   int     `json:"numMessages60days"`
            } `json:"spamStats"`
        } `json:"spamInfo"`
    } `json:"data"`
    Error int `json:"error"`
}