package models

// Application represents an application for which this authentication service controls access
type Application struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
