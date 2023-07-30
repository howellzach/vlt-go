package vlt

import (
	"encoding/json"
	"fmt"
)

type listSecretsResponse struct {
	Secrets []Secret `json:"secrets,omitempty"`
}

type secretResponse struct {
	Secret Secret `json:"secret,omitempty"`
}

type Secret struct {
	Name          string                 `json:"name,omitempty"`
	Version       Version                `json:"version,omitempty"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	LatestVersion int                    `json:"latest_version,string,omitempty"`
	CreatedBy     User                   `json:"created_by,omitempty"`
	SyncStatus    map[string]interface{} `json:"sync_status,omitempty"`
}

type Version struct {
	Version   int    `json:"version,string,omitempty"`
	Type      string `json:"type,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Value     string `json:"value,omitempty,omitempty"`
	CreatedBy User   `json:"created_by,omitempty"`
}

type User struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Email string `json:"email,omitempty"`
}

// GetSecret gets a secret object for a given secret name in an application
func (c *Client) GetSecret(name string) (Secret, error) {
	getSecretURL := BaseURL.JoinPath(
		fmt.Sprintf(
			"organizations/%s/projects/%s/apps/%s/open/%s",
			c.OrganizationID, c.ProjectID, c.ApplicationName, name,
		),
	)
	httpResponse, err := c.sendRequest("GET", *getSecretURL, nil, true)
	if err != nil {
		return Secret{}, err
	}

	var sResp secretResponse
	if err = json.Unmarshal(httpResponse, &sResp); err != nil {
		return Secret{}, err
	}

	return sResp.Secret, nil
}

// CreateSecret creates a new secret within an application
//
// If a secret already exists, a new version will be created/updated
// Only 20 versions of a secret can exist for the beta release of Vault Secrets,
// so an error will result from this function if a secret is found to be
// on its 20th version.
//
// If a secret needs to up dated past its 20th version, it will need to be deleted
// first.
func (c *Client) CreateSecret(name string, value string) (Secret, error) {
	secretLatestVersion, _ := c.GetLatestSecretVersion(name)
	if secretLatestVersion == 20 {
		return Secret{}, fmt.Errorf("Secret already has 20 versions - it will need to be deleted before creating a new version")
	}

	createSecretData := map[string]string{
		"name":  name,
		"value": value,
	}

	createSecretURL := BaseURL.JoinPath(
		fmt.Sprintf(
			"organizations/%s/projects/%s/apps/%s/kv",
			c.OrganizationID, c.ProjectID, c.ApplicationName,
		),
	)

	httpResponse, err := c.sendRequest("POST", *createSecretURL, createSecretData, true)
	if err != nil {
		return Secret{}, err
	}

	var createSecretResp secretResponse
	if err = json.Unmarshal(httpResponse, &createSecretResp); err != nil {
		return Secret{}, nil
	}

	return createSecretResp.Secret, nil
}

// GetLatestSecretVersion gets the latest version number for a given secret
//
// This is used for making sure CreateSecret is attempting to update a secret
// that already has 20 versions
func (c *Client) GetLatestSecretVersion(name string) (int, error) {
	secret, err := c.GetSecret(name)
	if err != nil {
		return 0, err
	}

	return secret.LatestVersion, nil
}

// ListSecrets gathers a list of secret names for a configured application
func (c *Client) ListSecrets() ([]string, error) {
	listSecretsURL := BaseURL.JoinPath(
		fmt.Sprintf(
			"organizations/%s/projects/%s/apps/%s/secrets",
			c.OrganizationID, c.ProjectID, c.ApplicationName,
		),
	)

	httpResponse, err := c.sendRequest("GET", *listSecretsURL, nil, true)
	if err != nil {
		return nil, err
	}

	var secretsResp listSecretsResponse
	if err = json.Unmarshal(httpResponse, &secretsResp); err != nil {
		return nil, err
	}

	var secretsList []string
	for _, secret := range secretsResp.Secrets {
		secretsList = append(secretsList, secret.Name)
	}

	return secretsList, nil
}

// GetAllSecrets gets all secret objects with their values in a configured application
func (c *Client) GetAllSecrets() ([]Secret, error) {
	secretList, err := c.ListSecrets()
	if err != nil {
		return nil, err
	}

	var secrets []Secret
	for _, secret := range secretList {
		s, err := c.GetSecret(secret)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}

	return secrets, nil
}

// DeleteSecret deletes a secret within an application by name
func (c *Client) DeleteSecret(name string) error {
	deleteSecretURL := BaseURL.JoinPath(
		fmt.Sprintf("organizations/%s/projects/%s/apps/%s/secrets/%s",
			c.OrganizationID, c.ProjectID, c.ApplicationName, name,
		),
	)

	_, err := c.sendRequest("DELETE", *deleteSecretURL, nil, true)
	if err != nil {
		return err
	}

	return nil
}
