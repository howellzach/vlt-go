# vlt-go

vlt-go is a Go client library for accessing the [HashiCorp Cloud Platform Vault Secrets Service](https://developer.hashicorp.com/hcp/docs/vault-secrets) released in beta in June 2023. The main advantage of Vault Secrets is that it does not require provisioning a full Vault cluster in order to utilize HashiCorp's Vault offering.

This library has been built to work with the `2023-06-13` version of the API and is a draft of what a full library for the API could look like.

This project was inspired by the SDK for Vault Secrets originally created by @ssbostan: [vault-secrets-sdk-go](https://github.com/ssbostan/vault-secrets-sdk-go)

## Limits and Constraints

HashiCorp has set certain [limits](https://developer.hashicorp.com/hcp/docs/vault-secrets/constraints-and-known-issues) in which this new service can be used in the public beta.
The main one for impacting this library is the limit of `20` versions for a secret.

# Installation

vlt-go is compatible with modern Go releases in module mode, with Go installed:
```
go get github.com/howellzach/vlt-go
```

# Usage

```go
import "github.com/howellzach/vlt-go"
```

```go
client, err := vlt.NewClient(
	os.Getenv("HCP_ORGANIZATION_ID"),
	os.Getenv("HCP_PROJECT_ID"), // optional, can be set to "" - it will be inferred from the project name, if project name is not set it will fetch the first project
	os.Getenv("HCP_PROJECT_NAME"), // ProjectID has priority over ProjectName
	os.Getenv("HCP_APPLICATION_NAME"),
	os.Getenv("HCP_CLIENT_ID"),
	os.Getenv("HCP_CLIENT_SECRET"),
)
if err != nil {
	log.Fatalln(err)
}

// get a secret object for stored secret named "TEST_SECRET"
secret, err := client.GetSecret("TEST_SECRET")
if err != nil {
	t.Errorf("client GetSecret error: %s", err.Error())
}

// access and use the value within the secret
secretValue := secret.Version.Value
fmt.Println("Secret value was %s", secretValue)
```

# Tests

All tests configured for this library are integration tests that actually perform real actions against a Vault Secrets application.
In order to test this library, the following secrets need to be present in an application.
- `TEST_SECRET1` : `this_is_a_value`
- `TEST_SECRET2` : `this_is_a_value_too`
- `TEST_SECRET3` : `created_secret`
