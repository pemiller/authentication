package datastore

import (
	"fmt"
	"strings"

	"pemiller/authentication/models"

	"github.com/couchbase/gocb"
)

// GetSite returns the site by ID
func (s *Store) GetSite(id string) (*models.Site, error) {
	key := s.GetSiteKey(id)

	var site models.Site

	_, err := s.bucket.Get(key, &site)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &site, nil
}

// GetSiteKey created a document key for a Site document
func (s *Store) GetSiteKey(id string) string {
	return fmt.Sprintf("site:%s", strings.ToLower(id))
}
