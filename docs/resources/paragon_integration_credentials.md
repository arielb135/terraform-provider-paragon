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

## Schema

### Argument Reference

- `integration_id` (String, Required) Identifier of the integration for which to create credentials.
- `project_id` (String, Required) Identifier of the project for which to create credentials.
- `oauth` (Object, Required) OAuth credentials for the relevant OAuth service.
  - `client_id` (String, Required) Client ID for the OAuth service.
  - `client_secret` (String, Required) Client secret for the OAuth service.
  - `scopes` (List of Strings, Optional) Scopes for the OAuth service, Please note per integration which are mandatory to avoid choosing incorrect scopes. should not be specified for custom integrations.

### Attributes Reference

- `id` (String) The unique identifier of the credentials resource.
- `creds_provider` (String) Provider of the credentials (e.g., "custom" for custom integration, "jira").
- `scheme` (String) The scheme used for authentication (e.g., "oauth_app").

## JSON State Structure Example

Here's a state sample, Please make sure you keep the `client_secret' attribute secured

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
