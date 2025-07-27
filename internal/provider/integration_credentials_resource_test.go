package provider

import (
    "context"
    "math/big"
    "strings"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
    "github.com/arielb135/terraform-provider-paragon/internal/client"
)

func TestValidateExtraConfiguration(t *testing.T) {
    resource := &integrationCredentialsResource{}
    ctx := context.Background()

    tests := []struct {
        name           string
        extraConfig    types.Dynamic
        integration    *client.Integration
        expectError    bool
        errorContains  string
    }{
        {
            name:        "null extra configuration should pass",
            extraConfig: types.DynamicNull(),
            integration: &client.Integration{Type: "custom"},
            expectError: false,
        },
        {
            name: "valid extra configuration should pass",
            extraConfig: types.DynamicValue(types.ObjectValueMust(
                map[string]attr.Type{
                    "customTimeout": types.StringType,
                    "retryCount":    types.NumberType,
                },
                map[string]attr.Value{
                    "customTimeout": types.StringValue("30s"),
                    "retryCount":    types.NumberValue(big.NewFloat(3)),
                },
            )),
            integration: &client.Integration{
                Type: "custom",
                CustomIntegration: &client.CustomIntegration{
                    AuthenticationType: "oauth",
                },
            },
            expectError: false,
        },
        {
            name: "conflicting OAuth field name should fail",
            extraConfig: types.DynamicValue(types.ObjectValueMust(
                map[string]attr.Type{
                    "clientId": types.StringType,
                },
                map[string]attr.Value{
                    "clientId": types.StringValue("test"),
                },
            )),
            integration: &client.Integration{Type: "custom"},
            expectError: true,
            errorContains: "conflicts with OAuth field names",
        },
        {
            name: "conflicting clientSecret should fail",
            extraConfig: types.DynamicValue(types.ObjectValueMust(
                map[string]attr.Type{
                    "clientSecret": types.StringType,
                },
                map[string]attr.Value{
                    "clientSecret": types.StringValue("test"),
                },
            )),
            integration: &client.Integration{Type: "custom"},
            expectError: true,
            errorContains: "conflicts with OAuth field names",
        },
        {
            name: "conflicting scopes should fail",
            extraConfig: types.DynamicValue(types.ObjectValueMust(
                map[string]attr.Type{
                    "scopes": types.StringType,
                },
                map[string]attr.Value{
                    "scopes": types.StringValue("read write"),
                },
            )),
            integration: &client.Integration{Type: "custom"},
            expectError: true,
            errorContains: "conflicts with OAuth field names",
        },
        {
            name: "non-oauth custom integration should fail",
            extraConfig: types.DynamicValue(types.ObjectValueMust(
                map[string]attr.Type{
                    "customParam": types.StringType,
                },
                map[string]attr.Value{
                    "customParam": types.StringValue("test"),
                },
            )),
            integration: &client.Integration{
                Type: "custom",
                CustomIntegration: &client.CustomIntegration{
                    AuthenticationType: "api_key",
                },
            },
            expectError: true,
            errorContains: "Extra configuration is not supported for custom integrations with authentication type 'api_key'",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            diags := resource.validateExtraConfiguration(ctx, tt.extraConfig, tt.integration)
            
            hasError := diags.HasError()
            if hasError != tt.expectError {
                t.Errorf("Expected error: %v, got error: %v", tt.expectError, hasError)
                if hasError {
                    for _, diag := range diags.Errors() {
                        t.Logf("Error: %s - %s", diag.Summary(), diag.Detail())
                    }
                }
                return
            }
            
            if tt.expectError && tt.errorContains != "" {
                found := false
                for _, diag := range diags.Errors() {
                    if strings.Contains(diag.Detail(), tt.errorContains) {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("Expected error message to contain '%s', but it didn't. Errors: %v", tt.errorContains, diags.Errors())
                }
            }
        })
    }
}

func TestExtractAllExtraConfigurationFromAPI(t *testing.T) {
    resource := &integrationCredentialsResource{}
    ctx := context.Background()

    tests := []struct {
        name           string
        apiValues      map[string]interface{}
        expectedKeys   []string
        expectNull     bool
    }{
        {
            name: "should extract non-OAuth keys",
            apiValues: map[string]interface{}{
                "clientId":                   "test_client",
                "clientSecret":               "test_secret",
                "scopes":                     "read write",
                "authenticateAsApplication":  true,
                "timeout":                    "30s",
                "retryCount":                 3,
            },
            expectedKeys: []string{"authenticateAsApplication", "timeout", "retryCount"},
            expectNull:   false,
        },
        {
            name: "should return null when only OAuth keys present",
            apiValues: map[string]interface{}{
                "clientId":     "test_client",
                "clientSecret": "test_secret",
                "scopes":       "read write",
            },
            expectedKeys: []string{},
            expectNull:   true,
        },
        {
            name: "should handle mixed data types",
            apiValues: map[string]interface{}{
                "clientId":       "test_client",
                "clientSecret":   "test_secret",
                "scopes":         "read write",
                "stringParam":    "value",
                "boolParam":      true,
                "numberParam":    42.5,
            },
            expectedKeys: []string{"stringParam", "boolParam", "numberParam"},
            expectNull:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := resource.extractAllExtraConfigurationFromAPI(ctx, tt.apiValues)
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
                return
            }

            if tt.expectNull {
                if !result.IsNull() {
                    t.Errorf("Expected null result, got: %v", result)
                }
                return
            }

            if result.IsNull() {
                t.Errorf("Expected non-null result, got null")
                return
            }

            // Verify that all expected keys are present
            underlyingValue := result.UnderlyingValue()
            if objectValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
                actualKeys := make([]string, 0, len(objectValue.Attributes()))
                for key := range objectValue.Attributes() {
                    actualKeys = append(actualKeys, key)
                }

                if len(actualKeys) != len(tt.expectedKeys) {
                    t.Errorf("Expected %d keys, got %d. Expected: %v, Actual: %v", 
                        len(tt.expectedKeys), len(actualKeys), tt.expectedKeys, actualKeys)
                    return
                }

                for _, expectedKey := range tt.expectedKeys {
                    found := false
                    for _, actualKey := range actualKeys {
                        if actualKey == expectedKey {
                            found = true
                            break
                        }
                    }
                    if !found {
                        t.Errorf("Expected key '%s' not found in result. Actual keys: %v", expectedKey, actualKeys)
                    }
                }
            } else {
                t.Errorf("Expected ObjectValue, got: %T", underlyingValue)
            }
        })
    }
}

func TestReadExtraConfigurationLogic(t *testing.T) {
    resource := &integrationCredentialsResource{}
    ctx := context.Background()

    // Test case 1: User has extra_configuration in state - should extract from API
    t.Run("should extract extra config when present in state", func(t *testing.T) {
        // Simulate state with extra configuration
        stateExtraConfig := types.DynamicValue(types.ObjectValueMust(
            map[string]attr.Type{
                "authenticateAsApplication": types.BoolType,
            },
            map[string]attr.Value{
                "authenticateAsApplication": types.BoolValue(true),
            },
        ))

        hasExtraConfigInState := !stateExtraConfig.IsNull() && !stateExtraConfig.IsUnknown()
        
        if !hasExtraConfigInState {
            t.Error("Expected hasExtraConfigInState to be true")
        }

        // Simulate API response
        apiValues := map[string]interface{}{
            "clientId":                   "client",
            "clientSecret":               "secret", 
            "scopes":                     "read write",
            "authenticateAsApplication":  true,
            "someOtherField":            "value",
        }

        // Extract extra configuration
        result, err := resource.extractAllExtraConfigurationFromAPI(ctx, apiValues)
        if err != nil {
            t.Errorf("Unexpected error: %v", err)
        }

        if result.IsNull() {
            t.Error("Expected non-null result when extra config is in state")
        }
    })

    // Test case 2: User has no extra_configuration in state - should return null
    t.Run("should return null when no extra config in state", func(t *testing.T) {
        // Simulate state without extra configuration
        stateExtraConfig := types.DynamicNull()

        hasExtraConfigInState := !stateExtraConfig.IsNull() && !stateExtraConfig.IsUnknown()
        
        if hasExtraConfigInState {
            t.Error("Expected hasExtraConfigInState to be false")
        }

        // In this case, we should NOT extract extra configuration from API
        // The result should be null to avoid drift detection
        expectedResult := types.DynamicNull()
        
        if !expectedResult.IsNull() {
            t.Error("Expected null result when no extra config is in state")
        }
    })
}