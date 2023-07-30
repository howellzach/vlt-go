package vlt

import (
	"net/http"
	"os"
	"testing"
)

var client = Client{
	OrganizationID:  os.Getenv("HCP_ORGANIZATION_ID"),
	ProjectID:       os.Getenv("HCP_PROJECT_ID"),
	ApplicationName: os.Getenv("HCP_APPLICATION_NAME"),
	ClientID:        os.Getenv("HCP_CLIENT_ID"),
	ClientSecret:    os.Getenv("HCP_CLIENT_SECRET"),
	httpClient:      &http.Client{},
}

func TestAuthenticate(t *testing.T) {
	if err := client.Authenticate(); err != nil {
		t.Fatalf("client Authenticate error: %s", err)
	}

	if client.AccessToken == "" {
		t.Fatalf("client Authenticate error: Could not retrieve access token.")
	}
}
