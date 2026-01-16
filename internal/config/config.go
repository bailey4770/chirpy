// Package config defines APIConfig struct type for admin api calls and db queries
package config

import (
	"github.com/bailey4770/chirpy/internal/database"
)

type APIConfig struct {
	DB *database.Queries
}

func New(dbQueries *database.Queries) *APIConfig {
	return &APIConfig{
		DB: dbQueries,
	}
}
