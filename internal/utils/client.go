package utils

import (
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/config"
)

func LoadClient() (api.Client, error) {
	token, err := config.LoadToken()
	if err != nil {
		return api.Client{}, fmt.Errorf("failed to load token. Please run 'tickli init' first: %w", err)
	}
	return *api.NewClient(token), nil
}
