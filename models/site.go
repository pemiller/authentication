package models

// Site represents subdivision of access within an application
type Site struct {
	SiteID     string `json:"site_id"`
	SiteName   string `json:"site_name"`
	SiteNumber string `json:"site_number"`
	SiteURL    string `json:"site_url"`
	IsActive   bool   `json:"is_active"`
}
