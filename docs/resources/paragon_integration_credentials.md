---
page_title: "integration_credentials Resource - paragon"
subcategory: ""
description: |-
  Manages credentials for an integration.
---

# paragon_integration_credentials (Resource)

Manages credentials for an integration.

~> **IMPORTANT:** 
The credentials should be stored securely and not exposed in any public repositories.

-> **NOTE:** Currently only oauth creds are supported, Trying to update non-oauth creds will result in an error

-> **NOTE:** For regular non-custom integration, there's no way verifying what type of authentication they required, so there's no restriction updating them.

-> **NOTE:** `scopes` should not be specified for custom integrations as they are specified in the custom builder.

## Scopes in oauth app
Oauth integrations usually come with default scopes that if not supplied - might cause the integration not to work.
It's highly recommended to check them out (via UI -> Settings -> so you can set them as a resource, for example - the basic jira configurations look like this:
```terraform
resource "paragon_integration_credentials" "jira" {
  integration_id = "d589fe10-b66e-4cb2-885a-0440393886f4"
  project_id = "6c9880c7-66af-467a-b319-0ce70e886bac"
  oauth = {
    client_id = "client_id"
    client_secret = "secret"
    scopes = ["offline_access", "read:jira-user"]
  }
}
```

## Example Usage

Use `paragon_integrations` data source to find out the relevant `integration_id`.

```terraform
# Create credentials for integrating a service
resource "paragon_integration_credentials" "example" {
  integration_id = "d589fe10-b66e-4cb2-885a-0440393886f4"
  project_id = "6c9880c7-66af-467a-b319-0ce70e886bac"
  oauth = {
    client_id = "client_id"
    client_secret = "secret"
    scopes = ["scope1", "scope2"]
  }
}
```

## Extra Configuration

The `extra_configuration` attribute allows you to specify additional configuration parameters beyond the standard OAuth fields. This is particularly useful for integrations that require custom parameters or advanced configuration options.

### Usage Examples

```terraform
# Example with extra configuration for custom timeout and retry settings
resource "paragon_integration_credentials" "custom_integration" {
  integration_id = "d589fe10-b66e-4cb2-885a-0440393886f4"
  project_id = "6c9880c7-66af-467a-b319-0ce70e886bac"
  oauth = {
    client_id = "client_id"
    client_secret = "secret"
  }
  # Note: extra_configuration is at the same level as oauth, not nested inside it
  extra_configuration = {
    timeout = "30s"
    retry_count = 3
    enable_debug = true
    custom_endpoint = "https://api.custom.com/v2"
  }
}

# Example with mixed data types in extra configuration
resource "paragon_integration_credentials" "advanced_config" {
  integration_id = "d589fe10-b66e-4cb2-885a-0440393886f4"
  project_id = "6c9880c7-66af-467a-b319-0ce70e886bac"
  oauth = {
    client_id = "client_id"
    client_secret = "secret"
    scopes = ["read", "write"]
  }
  # extra_configuration is a top-level attribute
  extra_configuration = {
    api_version = "v2"
    max_connections = 10
    use_ssl = true
    region = "us-east-1"
  }
}

# Real-world example with authenticateAsApplication
resource "paragon_integration_credentials" "jira_app_auth" {
  integration_id = "e09db889-dad3-49e4-895b-96b22d7de0db"
  project_id = local.project_id
  oauth = {
    client_id = "your_client_id"
    client_secret = "your_client_secret"
    scopes = ["comments:create", "issues:create", "read", "write"]
  }
  extra_configuration = {
    authenticateAsApplication = true
  }
}
```

### Important Notes

~> **IMPORTANT:** Extra configuration keys cannot conflict with OAuth field names (`clientId`, `clientSecret`, `scopes`). Using these reserved names will result in a validation error.

-> **NOTE:** Extra configuration is only supported for OAuth-based custom integrations. For custom integrations with other authentication types (e.g., API key), extra configuration is not allowed.

-> **NOTE:** For non-custom integrations, extra configuration can be used but should be applied carefully as it may interfere with standard integration behavior.

### Validation Rules

The provider validates extra configuration to ensure:

1. **No OAuth Field Conflicts**: Keys like `clientId`, `clientSecret`, and `scopes` are reserved for OAuth configuration
2. **Authentication Type Compatibility**: Extra configuration is only allowed for OAuth-based custom integrations
3. **Data Type Support**: Supports string, number, and boolean values in extra configuration

## Schema

### Argument Reference

- `integration_id` (String, Required) Identifier of the integration for which to create credentials.
- `project_id` (String, Required) Identifier of the project for which to create credentials.
- `oauth` (Object, Required) OAuth credentials for the relevant OAuth service.
  - `client_id` (String, Required) Client ID for the OAuth service.
  - `client_secret` (String, Required) Client secret for the OAuth service.
  - `scopes` (List of Strings, Optional) Scopes for the OAuth service, Please note per integration which are mandatory to avoid choosing incorrect scopes. should not be specified for custom integrations.
- `extra_configuration` (Dynamic, Optional, Sensitive) Additional configuration parameters for the integration credentials. Supports string, number, and boolean values. Cannot use reserved OAuth field names (`clientId`, `clientSecret`, `scopes`). Only supported for OAuth-based custom integrations.

### Attributes Reference

- `id` (String) The unique identifier of the credentials resource.
- `creds_provider` (String) Provider of the credentials (e.g., "custom" for custom integration, "jira").
- `scheme` (String) The scheme used for authentication (e.g., "oauth_app").

## JSON State Structure Example

Here's a state sample, Please make sure you keep the `client_secret` and `extra_configuration` attributes secured

```json
{
    "creds_provider": "custom",
    "id": "b9447451-56e0-4f70-a6df-2be85597e859",
    "integration_id": "d589fe10-b66e-4cb2-885a-0440393886f4",
    "project_id": "6c9880c7-66af-467a-b319-0ce70e886bac",
    "oauth": {
      "client_id": "your_client_id",
      "client_secret": "your_client_secret",
      "scopes": [
                "scope1",
                "scope2"
              ]
    },
    "extra_configuration": {
      "timeout": "30s",
      "retry_count": 3,
      "enable_debug": true,
      "api_version": "v2"
    },
    "scheme": "oauth_app"
}
```

### Basic State Example (without extra configuration)

```json
{
    "creds_provider": "jira",
    "id": "b9447451-56e0-4f70-a6df-2be85597e859",
    "integration_id": "d589fe10-b66e-4cb2-885a-0440393886f4",
    "project_id": "6c9880c7-66af-467a-b319-0ce70e886bac",
    "oauth": {
      "client_id": "your_client_id",
      "client_secret": "your_client_secret",
      "scopes": [
                "scope1",
                "scope2"
              ]
    },   
    "scheme": "oauth_app"
}
```
