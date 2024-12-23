// models/linkedin.go
package models

type LinkedInRequest struct {
    Name        string `json:"name"`
    CompanyName string `json:"company_name"`
    JobTitle    string `json:"job_title"`
    Location    string `json:"location"`
    Keywords    string `json:"keywords"`
    Limit       int    `json:"limit"`
}

type LinkedInEducation struct {
    Activities        string `json:"activities"`
    DateRange        string `json:"date_range"`
    Degree           string `json:"degree"`
    EndMonth         string `json:"end_month"`
    EndYear          int    `json:"end_year"`
    FieldOfStudy     string `json:"field_of_study"`
    School           string `json:"school"`
    SchoolID         string `json:"school_id"`
    SchoolLinkedinURL string `json:"school_linkedin_url"`
    SchoolLogoURL    string `json:"school_logo_url"`
    StartMonth       string `json:"start_month"`
    StartYear        int    `json:"start_year"`
}

type LinkedInExperience struct {
    Company            string `json:"company"`
    CompanyID          string `json:"company_id"`
    CompanyLinkedinURL string `json:"company_linkedin_url"`
    CompanyLogoURL     string `json:"company_logo_url"`
    DateRange          string `json:"date_range"`
    Description        string `json:"description"`
    Duration          string `json:"duration"`
    EndMonth          int    `json:"end_month"`
    EndYear           int    `json:"end_year"`
    IsCurrent         bool   `json:"is_current"`
    JobType           string `json:"job_type"`
    Location          string `json:"location"`
    Skills            string `json:"skills"`
    StartMonth        int    `json:"start_month"`
    StartYear         int    `json:"start_year"`
    Title             string `json:"title"`
}

type LinkedInProfile struct {
    MatchScore        int               `json:"_match_score"`
    About             string            `json:"about"`
    Company           string            `json:"company"`
    ConnectionCount   int               `json:"connection_count"`
    Country           string            `json:"country"`
    Educations        []LinkedInEducation `json:"educations"`
    Experiences       []LinkedInExperience `json:"experiences"`
    FirstName         string            `json:"first_name"`
    FollowerCount     int               `json:"follower_count"`
    FullName          string            `json:"full_name"`
    Headline          string            `json:"headline"`
    JobTitle          string            `json:"job_title"`
    Languages         string            `json:"languages"`
    LastName          string            `json:"last_name"`
    LinkedinURL       string            `json:"linkedin_url"`
    Location          string            `json:"location"`
    ProfileID         string            `json:"profile_id"`
    ProfileImageURL   string            `json:"profile_image_url"`
    School            string            `json:"school"`
}