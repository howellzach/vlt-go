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
	version       = "v1.1.0"
	AuthURL       = url.URL{Scheme: "https", Host: "auth.hashicorp.com", Path: "/oauth/token"}
	BaseURL       = url.URL{Scheme: "https", Host: "api.cloud.hashicorp.com", Path: "/secrets/2023-06-13"}
	HCPOrgURL     = url.URL{Scheme: "https", Host: "api.hashicorp.cloud", Path: "/resource-manager/2019-12-10/organizations"}
	HCPProjectURL = url.URL{Scheme: "https", Host: "api.hashicorp.cloud", Path: "/resource-manager/2019-12-10/projects"}
)

type Client struct {
	OrganizationID  string `json:"organization_id,omitempty"`
	ProjectID       string `json:"project_id,omitempty"`
	ProjectName     string `json:"project_name,omitempty"`
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

type organizationsResponse struct {
	Organizations []organization `json:"organizations,omitempty"`
}

type organization struct {
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Owner     map[string]interface{} `json:"owner,omitempty"`
	CreatedAt string                 `json:"created_at,omitempty"`
	State     string                 `json:"state,omitempty"`
}

type projectsResponse struct {
	Projects []project `json:"projects,omitempty"`
}

type project struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Parent      map[string]interface{} `json:"parent,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	State       string                 `json:"state,omitempty"`
	Description string                 `json:"description,omitempty"`
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

// Populate OrganizationID with the first organization found in the account
func (c *Client) PopulateOrganizationID() error {
	httpResponse, err := c.sendRequest("GET", HCPOrgURL, nil, true)
	if err != nil {
		return err
	}

	var orgResp organizationsResponse
	if err := json.Unmarshal(httpResponse, &orgResp); err != nil {
		return err
	}

	c.OrganizationID = orgResp.Organizations[0].ID
	return nil
}

// Populate ProjectID with the first project found in the organization or with the ID for the provided project name
func (c *Client) PopulateProjectID() error {
	q := HCPProjectURL.Query()
	q.Set("scope.type", "ORGANIZATION")
	q.Set("scope.id", c.OrganizationID)
	HCPProjectURL.RawQuery = q.Encode()

	httpResponse, err := c.sendRequest("GET", HCPProjectURL, nil, true)
	if err != nil {
		return err
	}

	var projectResp projectsResponse
	if err := json.Unmarshal(httpResponse, &projectResp); err != nil {
		return err
	}

	if len(projectResp.Projects) == 0 {
		return fmt.Errorf("No projects found in the organization")
	}

	if c.ProjectName == "" {
		c.ProjectID = projectResp.Projects[0].ID
		c.ProjectName = projectResp.Projects[0].Name
		return nil
	}

	for _, project := range projectResp.Projects {
		if c.ProjectName == project.Name {
			c.ProjectID = project.ID
			return nil
		}
	}

	return nil
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
func NewClient(organizationID string, projectID string, applicationName string, clientID string, clientSecret string, projectName string) (Client, error) {
	client := Client{
		OrganizationID:  organizationID,
		ProjectID:       projectID,
		ProjectName:     projectName,
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

	if client.OrganizationID == "" {
		if err := client.PopulateOrganizationID(); err != nil {
			return Client{}, err
		}
	}

	if client.ProjectID == "" {
		if err := client.PopulateProjectID(); err != nil {
			return Client{}, err
		}
	}

	return client, nil
}
