package vlt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	version = "v1.0.0"
	AuthURL = url.URL{Scheme: "https", Host: "auth.hashicorp.com", Path: "/oauth/token"}
	BaseURL = url.URL{Scheme: "https", Host: "api.cloud.hashicorp.com", Path: "/secrets/2023-06-13"}
)

type Client struct {
	OrganizationID  string
	ProjectID       string
	ApplicationName string
	ClientID        string
	ClientSecret    string
	AccessToken     string
	httpClient      *http.Client
}

type authResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
}

type authError struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type errorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (c *Client) sendRequest(method string, url url.URL, body interface{}, authed bool) ([]byte, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("vlt %s", version))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authed {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

	httpResponseData, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, handleHTTPError(httpResponseData, httpResponse, authed)
	}

	return httpResponseData, nil
}

func handleHTTPError(httpResponseData []byte, httpResponse *http.Response, authed bool) error {
	if authed {
		var errorResp errorResponse
		if err := json.Unmarshal(httpResponseData, &errorResp); err != nil {
			return err
		}
		return fmt.Errorf(
			fmt.Sprintf("VaultSecrets error: %s, HTTP status: %d", errorResp.Message, httpResponse.StatusCode),
		)
	} else {
		var errorResp authError
		if err := json.Unmarshal(httpResponseData, &errorResp); err != nil {
			return err
		}
		return fmt.Errorf(
			fmt.Sprintf("Authentication error: %s, HTTP status: %d", errorResp.ErrorDescription, httpResponse.StatusCode),
		)
	}
}

// Authenticate authenticates against HashiCorp's cloud services with provided client credentials
func (c *Client) Authenticate() error {
	authRequestData := map[string]string{
		"audience":      "https://api.hashicorp.cloud",
		"grant_type":    "client_credentials",
		"client_id":     c.ClientID,
		"client_secret": c.ClientSecret,
	}

	httpResponse, err := c.sendRequest("POST", AuthURL, authRequestData, false)
	if err != nil {
		return err
	}

	var authResp authResponse
	if err := json.Unmarshal(httpResponse, &authResp); err != nil {
		return err
	}

	c.AccessToken = authResp.AccessToken
	return nil
}

// NewClient creates a new client for interacting with the HashiCorp Vault Secrets service
func NewClient(organizationID string, projectID string, applicationName string, clientID string, clientSecret string) (Client, error) {
	client := Client{
		OrganizationID:  organizationID,
		ProjectID:       projectID,
		ApplicationName: applicationName,
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	if err := client.Authenticate(); err != nil {
		return Client{}, err
	}

	return client, nil
}
