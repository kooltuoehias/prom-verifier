package client

import (
	"fmt"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// NewAPI creates a new Prometheus v1 API client.
func NewAPI(url string) (v1.API, error) {
	client, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}
	return v1.NewAPI(client), nil
}
