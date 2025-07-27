# Design Document

## Overview

This feature adds support for an `extra_configuration` parameter to the `paragon_integration_credentials` resource. The parameter allows users to specify additional key-value pairs that get merged with the standard OAuth configuration when creating or updating integration credentials. The design maintains backward compatibility while extending the flexibility of the credential configuration system.

## Architecture

The implementation involves modifications to three main components:

1. **Terraform Resource Schema**: Add the `extra_configuration` attribute to the resource schema
2. **Client API Layer**: Modify the `CreateIntegrationCredentialsRequest` to support flexible value types
3. **State Management**: Implement logic to properly read back only the user-specified extra configuration keys

### Data Flow

```
User Configuration → Terraform Resource → Client API → Paragon API
                                                    ↓
User State ← Terraform Resource ← Client API ← Paragon API Response
```

## Components and Interfaces

### 1. Resource Schema Changes

The `integrationCredentialsResourceModel` will be extended with a new field:

```go
type integrationCredentialsResourceModel struct {
    ID                  types.String `tfsdk:"id"`
    ProjectID           types.String `tfsdk:"project_id"`
    IntegrationID       types.String `tfsdk:"integration_id"`
    Scheme              types.String `tfsdk:"scheme"`
    Provider            types.String `tfsdk:"creds_provider"`
    OAuth               *oauthModel  `tfsdk:"oauth"`
    ExtraConfiguration  types.Map    `tfsdk:"extra_configuration"` // New field
}
```

The schema will include:

```go
"extra_configuration": schema.MapAttribute{
    Description: "Additional configuration parameters for the integration credentials.",
    ElementType: types.StringType, // Will be changed to support dynamic types
    Optional:    true,
    Sensitive:   true,
}
```

### 2. Client API Modifications

The `CreateIntegrationCredentialsRequest` struct will be modified to support flexible value types:

```go
type CreateIntegrationCredentialsRequest struct {
    Name          string         `json:"name"`
    Values        map[string]any `json:"values"` // Changed from OAuthValues
    Provider      string         `json:"provider"`
    Scheme        string         `json:"scheme"`
    IntegrationID string         `json:"integrationId"`
}
```

A helper function will merge OAuth values with extra configuration:

```go
func mergeCredentialValues(oauth *oauthModel, extraConfig types.Map) (map[string]any, error) {
    values := make(map[string]any)
    
    // Add OAuth values
    values["clientId"] = oauth.ClientID.ValueString()
    values["clientSecret"] = oauth.ClientSecret.ValueString()
    if !oauth.Scopes.IsNull() {
        // Handle scopes conversion
    }
    
    // Add extra configuration
    if !extraConfig.IsNull() {
        extraConfigMap := make(map[string]attr.Value)
        extraConfig.ElementsAs(ctx, &extraConfigMap, false)
        
        for key, value := range extraConfigMap {
            // Convert terraform types to appropriate Go types
            values[key] = convertTerraformValue(value)
        }
    }
    
    return values, nil
}
```

### 3. State Management Logic

During the Read operation, the system will:

1. Retrieve the decrypted credential from the API
2. Extract only the keys that were originally specified in `extra_configuration`
3. Preserve the correct data types for each value
4. Update the state with the filtered extra configuration

```go
func extractExtraConfiguration(apiValues map[string]any, originalKeys []string) types.Map {
    filteredValues := make(map[string]attr.Value)
    
    for _, key := range originalKeys {
        if value, exists := apiValues[key]; exists {
            filteredValues[key] = convertToTerraformValue(value)
        }
    }
    
    return types.MapValueMust(types.StringType, filteredValues)
}
```

## Data Models

### Input Model (Terraform Configuration)

```hcl
resource "paragon_integration_credentials" "example" {
  integration_id = "integration-123"
  project_id     = "project-456"
  
  oauth = {
    client_id     = "client-id"
    client_secret = "client-secret"
    scopes        = ["read", "write"]
  }
  
  extra_configuration = {
    "authenticateAsApplication" = true
    "customTimeout"            = "30s"
    "retryCount"              = 3
  }
}
```

### API Request Model

```json
{
  "name": "user@example.com",
  "values": {
    "clientId": "client-id",
    "clientSecret": "client-secret",
    "scopes": "read write",
    "authenticateAsApplication": true,
    "customTimeout": "30s",
    "retryCount": 3
  },
  "provider": "custom",
  "scheme": "oauth_app",
  "integrationId": "integration-123"
}
```

### API Response Model

```json
{
  "id": "cred-789",
  "values": {
    "clientId": "client-id",
    "clientSecret": "client-secret",
    "scopes": "read write",
    "authenticateAsApplication": true,
    "customTimeout": "30s",
    "retryCount": 3,
    "accessTokenUrl": "https://api.example.com/oauth/token",
    "authUrl": "https://api.example.com/oauth/authorize"
  },
  "provider": "custom",
  "scheme": "oauth_app"
}
```

### State Model (Filtered)

Only the user-specified extra configuration keys are included in the state:

```go
ExtraConfiguration: types.MapValueMust(types.StringType, map[string]attr.Value{
    "authenticateAsApplication": types.BoolValue(true),
    "customTimeout":            types.StringValue("30s"),
    "retryCount":              types.NumberValue(3),
})
```

## Error Handling

### Validation Errors

1. **Type Validation**: Ensure extra configuration values are of supported types (string, bool, number)
2. **Key Conflicts**: Prevent extra configuration keys from conflicting with OAuth field names
3. **Integration Type Validation**: Validate that extra configuration is appropriate for the integration type

### Runtime Errors

1. **API Errors**: Handle cases where the API rejects extra configuration parameters
2. **State Sync Errors**: Handle cases where API response doesn't contain expected extra configuration
3. **Type Conversion Errors**: Handle failures when converting between Terraform and Go types

## Implementation Phases

### Phase 1: Client API Changes
- Modify `CreateIntegrationCredentialsRequest` to use `map[string]any`
- Implement value merging helper functions
- Add type conversion utilities

### Phase 2: Resource Schema Updates
- Add `extra_configuration` attribute to schema
- Update resource model struct
- Implement validation logic

### Phase 3: CRUD Operations
- Update Create operation to merge extra configuration
- Update Read operation to filter API response
- Update Update operation to handle extra configuration changes
- Ensure Delete operation works correctly