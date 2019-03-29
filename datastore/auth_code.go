package datastore

import (
	"fmt"

	"github.com/couchbase/gocb"

	"pemiller/authentication/config"
	"pemiller/authentication/models"
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

	s.DeleteAuthCodeResponseFromCache(code)

	_, err := s.bucket.Remove(key, 0)
	if err == gocb.ErrKeyNotFound {
		return nil
	}

	return err
}

// GetAuthCodeResponseFromCache returns the AuthCodeReponse defined by the code
func (s *Store) GetAuthCodeResponseFromCache(code string) (*models.AuthCodeResponse, error) {
	key := s.GetAuthCodeResponseKey(code)

	if cacheAuthCode, found := s.cache.Get(key); found {
		return cacheAuthCode.(*models.AuthCodeResponse), nil
	}

	return nil, nil
}

// UpsertAuthCodeResponseToCache adds the AuthCodeResponse if it does not exist, else it updates it.
// Useful to not have to build the application and site models when returning the response
// shortly after building it.
func (s *Store) UpsertAuthCodeResponseToCache(authCodeResponse *models.AuthCodeResponse) error {
	key := s.GetAuthCodeResponseKey(authCodeResponse.Code)

	s.cache.Set(key, authCodeResponse, cacheExpiration)

	return nil
}

// DeleteAuthCodeResponseFromCache deletes the AuthCode represented by the token from Cache
func (s *Store) DeleteAuthCodeResponseFromCache(code string) error {
	key := s.GetAuthCodeResponseKey(code)

	s.cache.Delete(key)

	return nil
}

// GetAuthCodeKey created a document key for an AuthCode document
func (s *Store) GetAuthCodeKey(id string) string {
	return fmt.Sprintf("%s:auth_code:%s", config.ServiceName, id)
}

// GetAuthCodeResponseKey created a document key for an AuthCodeResponse document
func (s *Store) GetAuthCodeResponseKey(id string) string {
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
