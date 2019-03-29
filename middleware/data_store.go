package middleware

import (
	"log"
	"pemiller/authentication/config"
	"pemiller/authentication/datastore"
	"time"

	"github.com/gin-gonic/gin"
	cache "github.com/patrickmn/go-cache"
)

// SetupDataStore creates a Store struct that contains pointers to the couchbase document store
// and the cache service and then puts it in the context
func SetupDataStore() gin.HandlerFunc {
	cache := cache.New(5*time.Minute, 10*time.Minute)

	return func(c *gin.Context) {
		store, err := datastore.NewStore(cache, config.ServiceName, config.CouchbaseConnection)
		if err != nil {
			log.Fatal(err)
		}

		c.Set(datastore.ContextKey, store)
		c.Next()
	}
}
