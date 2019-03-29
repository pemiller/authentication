package datastore

import (
	"fmt"

	"github.com/couchbase/gocb"
	cache "github.com/patrickmn/go-cache"

	"pemiller/authentication/config"
	"pemiller/authentication/models"
)

const (
	n1qlGetApplications = "SELECT b.* FROM $bucket b WHERE b.__type = 'application' ORDER BY b.name"
)

// GetApplicationsList returns a list of Applications
func (s *Store) GetApplicationsList() ([]*models.Application, error) {
	rows, err := s.ExecuteQuery(n1qlGetApplications, nil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*models.Application{}
	var app *models.Application
	for rows.Next(&app) {
		apps = append(apps, app)
	}
	return apps, err
}

// GetApplication returns the application defined by the id
func (s *Store) GetApplication(id string) (*models.Application, error) {
	key := s.GetApplicationKey(id)

	if cacheApp, found := s.cache.Get(key); found {
		return cacheApp.(*models.Application), nil
	}

	var app models.Application
	_, err := s.bucket.Get(key, &app)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.cache.Set(key, &app, cache.DefaultExpiration)
	return &app, nil
}

// UpsertApplication upserts the Application
func (s *Store) UpsertApplication(app *models.Application) error {
	key := s.GetApplicationKey(app.ID)
	_, err := s.bucket.Upsert(key, app, 0)
	return err
}

// DeleteApplication deletes the Application represented by the id
func (s *Store) DeleteApplication(id string) error {
	key := s.GetApplicationKey(id)
	_, err := s.bucket.Remove(key, 0)
	if err == gocb.ErrKeyNotFound {
		return nil
	}

	return err
}

// GetApplicationKey created a document key for an Application document
func (s *Store) GetApplicationKey(id string) string {
	return fmt.Sprintf("%s:application:%s", config.ServiceName, id)
}
