package client

import (
	"errors"
	"os"

	hatchet "github.com/hatchet-dev/hatchet/sdks/go"
)

func HatchetClient() (*hatchet.Client, error) {
	// check for HATCHET_CLIENT_TOKEN
	token := os.Getenv("HATCHET_CLIENT_TOKEN")
	if token == "" {
		return nil, errors.New("HATCHET_CLIENT_TOKEN is not set")
	}

	return hatchet.NewClient()
}
