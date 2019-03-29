package datastore

import (
	"fmt"

	"github.com/couchbase/gocb"

	"github.com/pemiller/authentication/config"
	"github.com/pemiller/authentication/models"
)

// GetAuthCode returns the AuthCode defined by the code
func (s *Store) GetAuthCode(code string) (*models.AuthCode, error) {
	key := s.GetAuthCodeKey(code)

	var authCode models.AuthCode

	_, err := s.bucket.Get(key, &authCode)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.bucket.Touch(key, 0, s.getAuthCodeExpiry(authCode.AuthType))
	return &authCode, nil
}

// UpsertAuthCode upserts the AuthCode object to the document store
func (s *Store) UpsertAuthCode(authCode *models.AuthCode) error {
	key := s.GetAuthCodeKey(authCode.Code)
	_, err := s.bucket.Upsert(key, authCode, s.getAuthCodeExpiry(authCode.AuthType))
	return err
}

// DeleteAuthCode deletes the AuthCode represented by the code
func (s *Store) DeleteAuthCode(code string) error {
	key := s.GetAuthCodeKey(code)

	s.DeleteAuthCodeDetailedFromCache(code)

	_, err := s.bucket.Remove(key, 0)
	if err == gocb.ErrKeyNotFound {
		return nil
	}

	return err
}

// GetAuthCodeDetailedFromCache returns the AuthCodeReponse defined by the code
func (s *Store) GetAuthCodeDetailedFromCache(code string) (*models.AuthCodeDetailed, error) {
	key := s.GetAuthCodeDetailedKey(code)

	if cacheAuthCode, found := s.cache.Get(key); found {
		return cacheAuthCode.(*models.AuthCodeDetailed), nil
	}

	return nil, nil
}

// UpsertAuthCodeDetailedToCache adds the AuthCodeDetailed if it does not exist, else it updates it.
// Useful to not have to build the application and site models when returning the response
// shortly after building it.
func (s *Store) UpsertAuthCodeDetailedToCache(AuthCodeDetailed *models.AuthCodeDetailed) error {
	key := s.GetAuthCodeDetailedKey(AuthCodeDetailed.Code)

	s.cache.Set(key, AuthCodeDetailed, cacheExpiration)

	return nil
}

// DeleteAuthCodeDetailedFromCache deletes the AuthCode represented by the code from Cache
func (s *Store) DeleteAuthCodeDetailedFromCache(code string) error {
	key := s.GetAuthCodeDetailedKey(code)

	s.cache.Delete(key)

	return nil
}

// GetAuthCodeKey created a document key for an AuthCode document
func (s *Store) GetAuthCodeKey(id string) string {
	return fmt.Sprintf("%s:auth_code:%s", config.ServiceName, id)
}

// GetAuthCodeDetailedKey created a document key for an AuthCodeDetailed document
func (s *Store) GetAuthCodeDetailedKey(id string) string {
	return fmt.Sprintf("%s:auth_code_response:%s", config.ServiceName, id)
}

func (s *Store) getAuthCodeExpiry(authType models.AuthTypeValue) uint32 {
	if authType == models.AuthTypeUser {
		return authCodeExpiration
	}

	if authType == models.AuthTypeApplication {
		return applicationTokenExpiration
	}

	return 1
}
