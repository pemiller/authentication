package config

import (
	"fmt"

	"github.com/ianschenck/envflag"
)

// global configs
var (
	Port                string
	Address             string
	CouchbaseConnection string
)

// ServiceName ...
const ServiceName = "authentication"

func init() {
	envflag.StringVar(&Port, "PORT", "9199", "")
	envflag.StringVar(&CouchbaseConnection, "CB_CONNECTION", "", "")
}

// Parse loads and validates the config
func Parse() {
	envflag.Parse()
	if CouchbaseConnection == "" {
		panic("CB_CONNECTION missing")
	}

	Address = fmt.Sprintf(":%s", Port)
}
