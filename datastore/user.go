package datastore

import (
	"fmt"
	"strings"
	"time"

	"pemiller/authentication/config"
	"pemiller/authentication/models"

	"github.com/couchbase/gocb"
	gocbcore "gopkg.in/couchbase/gocbcore.v7"
)

// GetUser returns the user by ID
func (s *Store) GetUser(id string) (*models.User, error) {
	key := s.GetUserKey(id)
	return s.GetUserByKey(key)
}

// GetUserByKey returns the user by key
func (s *Store) GetUserByKey(key string) (*models.User, error) {
	var user models.User

	_, err := s.bucket.Get(key, &user)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, err
}

// GetUserByEmail returns the user by email
func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	userRef, err := s.GetUserRef(email)
	if err != nil {
		return nil, err
	}
	if userRef == nil {
		return nil, nil
	}

	var user models.User

	_, err = s.bucket.Get(userRef.UserRef, &user)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, err
}

// GetUserRef returns the UserRef document for the email
func (s *Store) GetUserRef(email string) (*models.UserRef, error) {
	key := s.GetUserRefKey(email)

	var userRef models.UserRef

	_, err := s.bucket.Get(key, &userRef)
	if err == gocb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userRef, err
}

// UserIsLocked returns true if account is locked
func (s *Store) UserIsLocked(email string) (bool, error) {
	key := s.GetLockedKey(email)

	var locked bool

	_, err := s.bucket.Get(key, &locked)
	if err == gocb.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}

	return locked, nil
}

// IncrLoginFailCount increments the fail count of a user and locks account when failed 5 times. Returns if locked
func (s *Store) IncrLoginFailCount(email string) (bool, error) {
	key := s.GetFailCountKey(email)

	i, _, err := s.bucket.Counter(key, 1, 1, failCountExpiration)
	if err != nil {
		return false, err
	}

	if i >= 5 {
		key = s.GetLockedKey(email)

		_, err := s.bucket.Upsert(key, true, failCountExpiration)
		if err != nil {
			return true, err
		}
	}

	return i > 5, nil
}

// ClearLoginFailCount clears failcount entry for the email
func (s *Store) ClearLoginFailCount(email string) error {
	key := s.GetFailCountKey(email)

	_, err := s.bucket.Remove(key, 0)
	if err == gocb.ErrKeyNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	key = s.GetLockedKey(email)

	_, err = s.bucket.Remove(key, 0)
	if err == gocb.ErrKeyNotFound {
		return nil
	}

	return err
}

// UpdateLoginDateForAll sets the Last Login fields for the ALL record
func (s *Store) UpdateLoginDateForAll(id string, authType models.AuthTypeValue, ip string) error {
	return s.updateLoginDate(id, "logins", authType, ip)
}

// GetUserKey created a document key for a User document
func (s *Store) GetUserKey(id string) string {
	return fmt.Sprintf("%s:user:%s", config.ServiceName, strings.ToLower(id))
}

// GetUserRefKey created a document key for a UserRef document
func (s *Store) GetUserRefKey(email string) string {
	return fmt.Sprintf("%s:user_ref:%s", config.ServiceName, strings.ToLower(email))
}

// GetLockedKey created a document key for a Locked document
func (s *Store) GetLockedKey(email string) string {
	return fmt.Sprintf("%s:locked:%s", config.ServiceName, strings.ToLower(email))
}

// GetFailCountKey created a document key for a FailCount document
func (s *Store) GetFailCountKey(email string) string {
	return fmt.Sprintf("%s:fail_count:%s", config.ServiceName, strings.ToLower(email))
}

func (s *Store) updateLoginDate(id, path string, authType models.AuthTypeValue, ip string) error {
	key := s.GetUserKey(id)

	frag, err := s.bucket.LookupIn(key).Get(path).Execute()
	if err != nil {
		if !gocbcore.IsErrorStatus(err, gocbcore.StatusSubDocPathNotFound) &&
			!gocbcore.IsErrorStatus(err, gocbcore.StatusSubDocBadMulti) {
			return err
		}
	}

	var logins []*models.LoginTime
	err = frag.Content("logins", &logins)
	if err != nil {
		if !gocbcore.IsErrorStatus(err, gocbcore.StatusSubDocPathNotFound) {
			return err
		}
	}

	newLogin := &models.LoginTime{
		AuthType: authType,
		Time:     time.Now().UTC(),
		IP:       ip,
	}
	logins = append([]*models.LoginTime{newLogin}, logins...)
	if len(logins) > 40 {
		logins = logins[0:40]
	}

	_, err = s.bucket.MutateIn(key, 0, 0).Upsert(path, logins, true).Execute()

	return err
}
