package datastore

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/couchbase/gocb"
	cache "github.com/patrickmn/go-cache"
)

// ContextKey is used to place Store in Context
const ContextKey = "datastore"

var (
	authCodeExpiration         = uint32((time.Hour * 336).Seconds()) // 2 weeks
	applicationTokenExpiration = uint32((time.Hour * 24).Seconds())  // 1 day
	accessTokenExpiration      = uint32((time.Hour * 24).Seconds())  // 1 day
	failCountExpiration        = uint32((time.Minute * 5).Seconds()) // 5 minutes
	cacheExpiration            = time.Duration(20 * time.Second)     // 20 seconds
)

// Store is an object that contains connections to data stores.
type Store struct {
	cluster    *gocb.Cluster
	bucket     *gocb.Bucket
	cache      *cache.Cache
	bucketName string
}

// NewStore initializes a Store with a connection to document and cache
func NewStore(cache *cache.Cache, service, source string) (*Store, error) {
	parsedURL, err := url.Parse(source)
	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme != "couchbase" {
		return nil, fmt.Errorf("invalid storage provider (%s)", parsedURL.Scheme)
	}

	password := ""
	user := ""
	if parsedURL.User != nil {
		password, _ = parsedURL.User.Password()
		user = parsedURL.User.Username()
	}

	if parsedURL.Path == "" || parsedURL.Path == "/" {
		return nil, fmt.Errorf("invalid bucket (%s)", parsedURL.Path)
	}

	spec := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	cluster, err := gocb.Connect(spec)
	if err != nil {
		return nil, fmt.Errorf("error initializing cluster: (%v)", err)
	}

	var bucket *gocb.Bucket
	if user != "" {
		cluster.Authenticate(gocb.PasswordAuthenticator{Username: user, Password: password})
		bucket, err = cluster.OpenBucket(parsedURL.Path[1:], "")
	} else {
		bucket, err = cluster.OpenBucket(parsedURL.Path[1:], password)
	}

	if err != nil {
		return nil, fmt.Errorf("error opening bucket: (%v)", err)
	}

	transcoder := DocTypeTranscoder{
		DefaultTranscoder: gocb.DefaultTranscoder{},
	}
	bucket.SetTranscoder(transcoder)

	return &Store{
		cluster:    cluster,
		bucket:     bucket,
		cache:      cache,
		bucketName: parsedURL.Path[1:],
	}, nil
}

// GetFromContext returns the Store associated with the context
func GetFromContext(c context.Context) *Store {
	s, _ := c.Value(ContextKey).(*Store)
	return s
}

// ExecuteQuery prepares a N1QL query for the current bucket, executes it and returns the results
func (s *Store) ExecuteQuery(query string, params interface{}, options ...func(*gocb.N1qlQuery)) (gocb.QueryResults, error) {
	if !strings.Contains(query, "$bucket") {
		return nil, errors.New("query does not contain the '$bucket' placeholder")
	}

	bucket := fmt.Sprintf("`%s`", s.bucketName)
	preparedQuery := strings.Replace(query, "$bucket", bucket, -1)

	n1ql := gocb.NewN1qlQuery(preparedQuery)
	for _, opt := range options {
		opt(n1ql)
	}

	return s.bucket.ExecuteN1qlQuery(n1ql, params)
}
