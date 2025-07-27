# Requirements Document

## Introduction

This feature adds support for an `extra_configuration` parameter to the `paragon_integration_credentials` resource. This parameter will allow users to specify additional key-value pairs that get merged into the credential values during creation and updates. The feature enables more flexible configuration of integration credentials by supporting custom parameters beyond the standard OAuth fields.

## Requirements

### Requirement 1

**User Story:** As a Terraform user, I want to specify additional configuration parameters for integration credentials, so that I can customize the behavior of integrations beyond the standard OAuth fields.

#### Acceptance Criteria

1. WHEN a user defines an `extra_configuration` block in the resource THEN the system SHALL accept a map of string keys to any value type (string, boolean, number)
2. WHEN creating integration credentials THEN the system SHALL merge the extra configuration values with the OAuth values in the API request
3. WHEN the extra configuration contains a boolean value THEN the system SHALL preserve the boolean type in the API request
4. WHEN the extra configuration contains a string value THEN the system SHALL preserve the string type in the API request
5. WHEN the extra configuration is not specified THEN the system SHALL function normally without any extra parameters

### Requirement 2

**User Story:** As a Terraform user, I want the extra configuration to be properly read back from the API, so that my Terraform state accurately reflects the deployed configuration.

#### Acceptance Criteria

1. WHEN reading credentials from the API THEN the system SHALL extract only the extra configuration keys that were originally specified in the resource
2. WHEN the API response contains additional system-generated fields THEN the system SHALL NOT include them in the extra_configuration state
3. WHEN an extra configuration value is a boolean in the API response THEN the system SHALL preserve it as a boolean in the state
4. WHEN an extra configuration value is a string in the API response THEN the system SHALL preserve it as a string in the state
5. WHEN the original resource had no extra_configuration THEN the system SHALL NOT populate extra_configuration in the state

### Requirement 3

**User Story:** As a Terraform user, I want updates to extra configuration to be properly handled, so that I can modify integration behavior through Terraform.

#### Acceptance Criteria

1. WHEN a user updates the extra_configuration block THEN the system SHALL send the updated values in the API request
2. WHEN a user removes a key from extra_configuration THEN the system SHALL not include that key in subsequent API requests
3. WHEN a user adds a new key to extra_configuration THEN the system SHALL include the new key in the API request
4. WHEN updating credentials THEN the system SHALL preserve the correct data types for all extra configuration values

### Requirement 4

**User Story:** As a developer, I want the client API to support flexible value types, so that the extra configuration feature can handle various data types.

#### Acceptance Criteria

1. WHEN the client sends credential values THEN the Values field SHALL support map[string]any type
2. WHEN processing OAuth values THEN the system SHALL maintain backward compatibility with existing OAuthValues structure
3. WHEN merging extra configuration THEN the system SHALL properly combine OAuth fields with additional parameters
4. WHEN serializing to JSON THEN the system SHALL preserve the correct data types for all values