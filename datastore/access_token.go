package datastore

import (
	"fmt"

	"github.com/couchbase/gocb"

	"github.com/pemiller/authentication/config"
	"github.com/pemiller/authentication/models"
)

// GetAccessToken returns the AccessToken defined by the token
func (s *Store) GetAccessToken(token string) (*models.AccessToken, error) {
	key := s.GetAccessTokenKey(token)

	var accessToken models.AccessToken

	_, err := s.bucket.GetAndTouch(key, accessTokenExpiration, &accessToken)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &accessToken, nil
}

// UpsertAccessToken upserts the AccessToken object to the document store
func (s *Store) UpsertAccessToken(accessToken *models.AccessToken) error {
	key := s.GetAccessTokenKey(accessToken.Token)
	_, err := s.bucket.Upsert(key, accessToken, accessTokenExpiration)
	return err
}

// DeleteAccessToken deletes the AccessToken represented by the token
func (s *Store) DeleteAccessToken(token string) error {
	key := s.GetAccessTokenKey(token)

	s.DeleteAccessTokenDetailedFromCache(token)

	_, err := s.bucket.Remove(key, 0)
	if err == gocb.ErrKeyNotFound {
		return nil
	}

	return err
}

// GetAccessTokenDetailedFromCache returns the AccessTokenReponse defined by the token
func (s *Store) GetAccessTokenDetailedFromCache(token string) (*models.AccessTokenDetailed, error) {
	key := s.GetAccessTokenDetailedKey(token)

	if cacheAccessToken, found := s.cache.Get(key); found {
		return cacheAccessToken.(*models.AccessTokenDetailed), nil
	}

	return nil, nil
}

// UpsertAccessTokenDetailedToCache adds the AccessTokenDetailed if it does not exist, else it updates it.
// Useful to not have to build the application and site models when returning the response
// shortly after building it.
func (s *Store) UpsertAccessTokenDetailedToCache(AccessTokenDetailed *models.AccessTokenDetailed) error {
	key := s.GetAccessTokenDetailedKey(AccessTokenDetailed.Token)

	s.cache.Set(key, AccessTokenDetailed, cacheExpiration)

	return nil
}

// DeleteAccessTokenDetailedFromCache deletes the AccessToken represented by the token from Cache
func (s *Store) DeleteAccessTokenDetailedFromCache(token string) error {
	key := s.GetAccessTokenDetailedKey(token)

	s.cache.Delete(key)

	return nil
}

// GetAccessTokenKey created a document key for an AccessToken document
func (s *Store) GetAccessTokenKey(token string) string {
	return fmt.Sprintf("%s:access_token:%s", config.ServiceName, token)
}

// GetAccessTokenDetailedKey created a document key for an AccessTokenDetailed document
func (s *Store) GetAccessTokenDetailedKey(token string) string {
	return fmt.Sprintf("%s:access_token_response:%s", config.ServiceName, token)
}
