package config

import (
	"fmt"
	"os"
)

// global configs
var (
	Port                string
	Address             string
	CouchbaseConnection string
)

// ServiceName ...
const ServiceName = "authentication"

// Parse loads and validates the config
func Parse() {
	Port = os.Getenv("AUTHENTICATION_PORT")
	CouchbaseConnection = os.Getenv("AUTHENTICATION_CB_CONNECTION")

	if Port == "" {
		panic("Port missing")
	}

	if CouchbaseConnection == "" {
		panic("Couchbase connection string missing")
	}

	Address = fmt.Sprintf(":%s", Port)
}
