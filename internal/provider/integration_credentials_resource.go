package provider

import (
    "context"
    "fmt"
    "strings"
    "math/big"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/arielb135/terraform-provider-paragon/internal/client"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-log/tflog"

)

// Ensure the implementation satisfies the expected interfaces.
var (
    _ resource.Resource              = &integrationCredentialsResource{}
    _ resource.ResourceWithConfigure = &integrationCredentialsResource{}
)

// NewIntegrationCredentialsResource is a helper function to simplify the provider implementation.
func NewIntegrationCredentialsResource() resource.Resource {
    return &integrationCredentialsResource{}
}

// integrationCredentialsResource is the resource implementation.
type integrationCredentialsResource struct {
    client *client.Client
}

// integrationCredentialsResourceModel maps the resource schema data.
type integrationCredentialsResourceModel struct {
    ID                 types.String  `tfsdk:"id"`
    ProjectID          types.String  `tfsdk:"project_id"`
    IntegrationID      types.String  `tfsdk:"integration_id"`
    Scheme             types.String  `tfsdk:"scheme"`
    Provider           types.String  `tfsdk:"creds_provider"`
    OAuth              *oauthModel   `tfsdk:"oauth"`
    ExtraConfiguration types.Dynamic `tfsdk:"extra_configuration"`
}

type oauthModel struct {
    ClientID     types.String `tfsdk:"client_id"`
    ClientSecret types.String `tfsdk:"client_secret"`
    Scopes       types.List   `tfsdk:"scopes"`
}

// Configure adds the provider configured client to the resource.
func (r *integrationCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    r.client = req.ProviderData.(*client.Client)
}

// Metadata returns the resource type name.
func (r *integrationCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_integration_credentials"
}

// Schema defines the schema for the resource.
func (r *integrationCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manages integration credentials.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "Identifier of the integration credentials.",
                Computed:    true,
            },
            "project_id": schema.StringAttribute{
                Description: "Identifier of the project.",
                Required:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "integration_id": schema.StringAttribute{
                Description: "Identifier of the integration.",
                Required:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "scheme": schema.StringAttribute{
                Description: "Scheme of the integration credentials.",
                Computed:    true,
            },
            "creds_provider": schema.StringAttribute{
                Description: "Provider of the integration credentials.",
                Computed:    true,
            },
            "oauth": schema.SingleNestedAttribute{
                Description: "OAuth configuration for the integration credentials.",
                Required:    true,
                Sensitive:   true,
                Attributes: map[string]schema.Attribute{
                    "client_id": schema.StringAttribute{
                        Description: "Client ID for OAuth.",
                        Required:    true,
                        Sensitive:   true,
                        Validators: []validator.String{
                            stringvalidator.LengthAtLeast(1),
                        },

                    },
                    "client_secret": schema.StringAttribute{
                        Description: "Client secret for OAuth.",
                        Required:    true,
                        Sensitive:   true,
                        Validators: []validator.String{
                            stringvalidator.LengthAtLeast(1),
                        },
                    },
                    "scopes": schema.ListAttribute{
                        Description: "Scopes for OAuth.",
                        ElementType: types.StringType,
                        Optional:    true,
                        Sensitive:   true,
                        Validators: []validator.List{   // Change from []validator.String to []validator.List
                            listvalidator.SizeAtLeast(1),
                        },
                    },
                },
            },
            "extra_configuration": schema.DynamicAttribute{
                Description: "Additional configuration parameters for the integration credentials.",
                Optional:    true,
                Sensitive:   true,
            },
        },
    }
}

// validateExtraConfiguration validates extra configuration against OAuth field conflicts and provider-specific rules
func (r *integrationCredentialsResource) validateExtraConfiguration(ctx context.Context, extraConfig types.Dynamic, integration *client.Integration) diag.Diagnostics {
    var diags diag.Diagnostics
    
    if extraConfig.IsNull() || extraConfig.IsUnknown() {
        return diags
    }
    
    // Define OAuth field names that cannot be used in extra configuration
    oauthFieldNames := map[string]bool{
        "clientId":     true,
        "clientSecret": true,
        "scopes":       true,
    }
    
    // Get the underlying value from the dynamic type
    underlyingValue := extraConfig.UnderlyingValue()
    
    // Convert to object value and validate the keys
    if objectValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
        for key := range objectValue.Attributes() {
            // Check for OAuth field name conflicts
            if oauthFieldNames[key] {
                diags.AddError(
                    "Invalid extra configuration key",
                    fmt.Sprintf("Extra configuration key '%s' conflicts with OAuth field names. "+
                        "The following keys are reserved for OAuth configuration: clientId, clientSecret, scopes", key),
                )
            }
        }
        
        // Provider-specific validation
        if integration != nil {
            if integration.Type == "custom" {
                // For custom integrations, validate that extra configuration is appropriate
                // Custom integrations may have more flexibility with extra configuration
                // but should still follow basic validation rules
                if integration.CustomIntegration != nil && integration.CustomIntegration.AuthenticationType != "oauth" {
                    diags.AddError(
                        "Invalid extra configuration for custom integration",
                        fmt.Sprintf("Extra configuration is not supported for custom integrations with authentication type '%s'. "+
                            "Extra configuration is only supported for OAuth-based custom integrations.", 
                            integration.CustomIntegration.AuthenticationType),
                    )
                }
            } else {
                // For non-custom integrations, extra configuration should be used carefully
                // as it may interfere with the standard integration behavior
                // This is more of a warning/informational validation
                if len(objectValue.Attributes()) > 0 {
                    // Allow extra configuration for non-custom integrations but ensure it's documented
                    // that users should be careful with these settings
                }
            }
        }
    }
    
    return diags
}

// getExtraConfigurationKeys extracts the keys from the extra configuration for filtering during reads
func (r *integrationCredentialsResource) getExtraConfigurationKeys(extraConfig types.Dynamic) []string {
    var keys []string
    
    if extraConfig.IsNull() || extraConfig.IsUnknown() {
        return keys
    }
    
    // Get the underlying value from the dynamic type
    underlyingValue := extraConfig.UnderlyingValue()
    
    // Convert to object value and extract the keys
    if objectValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
        for key := range objectValue.Attributes() {
            keys = append(keys, key)
        }
    }
    
    return keys
}

// convertAPIValueToTerraformValue converts API response values to Terraform attribute values
func (r *integrationCredentialsResource) convertAPIValueToTerraformValue(value interface{}) attr.Value {
    switch v := value.(type) {
    case string:
        return types.StringValue(v)
    case bool:
        return types.BoolValue(v)
    case float64:
        return types.NumberValue(big.NewFloat(v))
    case int:
        return types.NumberValue(big.NewFloat(float64(v)))
    case int64:
        return types.NumberValue(big.NewFloat(float64(v)))
    default:
        // For unknown types, convert to string as fallback
        return types.StringValue(fmt.Sprintf("%v", v))
    }
}

// extractExtraConfigurationFromAPI filters API response to include only user-specified extra configuration keys
func (r *integrationCredentialsResource) extractExtraConfigurationFromAPI(ctx context.Context, apiValues map[string]interface{}, originalKeys []string) (types.Dynamic, error) {
    if len(originalKeys) == 0 {
        return types.DynamicNull(), nil
    }
    
    filteredAttributes := make(map[string]attr.Value)
    attributeTypes := make(map[string]attr.Type)
    
    for _, key := range originalKeys {
        if value, exists := apiValues[key]; exists {
            terraformValue := r.convertAPIValueToTerraformValue(value)
            filteredAttributes[key] = terraformValue
            attributeTypes[key] = terraformValue.Type(ctx)
        }
    }
    
    if len(filteredAttributes) == 0 {
        return types.DynamicNull(), nil
    }
    
    // Create object value
    objectValue, diags := types.ObjectValue(attributeTypes, filteredAttributes)
    if diags.HasError() {
        return types.DynamicNull(), fmt.Errorf("failed to create object value: %v", diags)
    }
    
    // Convert to dynamic value
    dynamicValue := types.DynamicValue(objectValue)
    
    return dynamicValue, nil
}

// extractAllExtraConfigurationFromAPI extracts all non-OAuth keys from API response as extra configuration
func (r *integrationCredentialsResource) extractAllExtraConfigurationFromAPI(ctx context.Context, apiValues map[string]interface{}) (types.Dynamic, error) {
    // Define OAuth field names that should be excluded from extra configuration
    oauthFieldNames := map[string]bool{
        "clientId":     true,
        "clientSecret": true,
        "scopes":       true,
    }
    
    filteredAttributes := make(map[string]attr.Value)
    attributeTypes := make(map[string]attr.Type)
    
    // Extract all keys that are not OAuth fields
    for key, value := range apiValues {
        if !oauthFieldNames[key] {
            terraformValue := r.convertAPIValueToTerraformValue(value)
            filteredAttributes[key] = terraformValue
            attributeTypes[key] = terraformValue.Type(ctx)
        }
    }
    
    if len(filteredAttributes) == 0 {
        return types.DynamicNull(), nil
    }
    
    // Create object value
    objectValue, diags := types.ObjectValue(attributeTypes, filteredAttributes)
    if diags.HasError() {
        return types.DynamicNull(), fmt.Errorf("failed to create object value: %v", diags)
    }
    
    // Convert to dynamic value
    dynamicValue := types.DynamicValue(objectValue)
    
    return dynamicValue, nil
}

// mergeCredentialValues merges OAuth values with extra configuration values
func (r *integrationCredentialsResource) mergeCredentialValues(ctx context.Context, oauth *oauthModel, extraConfig types.Dynamic, scopesStr string) (map[string]any, error) {
    values := map[string]any{
        "clientId":     oauth.ClientID.ValueString(),
        "clientSecret": oauth.ClientSecret.ValueString(),
        "scopes":       scopesStr,
    }

    // If extra configuration is provided, merge it with OAuth values
    if !extraConfig.IsNull() && !extraConfig.IsUnknown() {
        // Get the underlying value from the dynamic type
        underlyingValue := extraConfig.UnderlyingValue()
        
        // Convert to object value and extract the attributes
        if objectValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
            for key, attrValue := range objectValue.Attributes() {
                // Convert terraform attribute values to Go values
                switch v := attrValue.(type) {
                case basetypes.StringValue:
                    values[key] = v.ValueString()
                case basetypes.BoolValue:
                    values[key] = v.ValueBool()
                case basetypes.NumberValue:
                    if v.ValueBigFloat() != nil {
                        if floatVal, accuracy := v.ValueBigFloat().Float64(); accuracy == 0 {
                            values[key] = floatVal
                        }
                    }
                default:
                    // For other types, try to get the underlying Go value
                    values[key] = attrValue
                }
            }
        }
    }

    return values, nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *integrationCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan integrationCredentialsResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := plan.ProjectID.ValueString()
    integrationID := plan.IntegrationID.ValueString()

    // Retrieve the integration
    integration, err := r.client.GetIntegration(ctx, projectID, integrationID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error retrieving integration",
            "Could not retrieve integration, unexpected error: "+err.Error(),
        )
        return
    }

    // Validate extra configuration
    validationDiags := r.validateExtraConfiguration(ctx, plan.ExtraConfiguration, integration)
    resp.Diagnostics.Append(validationDiags...)
    if resp.Diagnostics.HasError() {
        return
    }

    scopesStr := ""

    if integration.Type == "custom" {
        if plan.OAuth != nil && integration.CustomIntegration.AuthenticationType != "oauth" {
            resp.Diagnostics.AddError(
                "Invalid authentication type",
                "The 'oauth' block is specified, but the custom integration's authentication type is not 'oauth'",
            )
            return
        }

        if plan.OAuth != nil && !plan.OAuth.Scopes.IsNull() {
            resp.Diagnostics.AddError(
                "Unexpected scopes section",
                "Scopes cannot be specified for custom integrations",
            )
            return
        }
    } else {
        if plan.OAuth.Scopes.IsNull() {
            resp.Diagnostics.AddError(
                "Missing scopes",
                "Scopes must be specified for this integration",
            )
            return
        }

        // Extract the scopes
        var scopes []string
        diags = plan.OAuth.Scopes.ElementsAs(ctx, &scopes, false)
        resp.Diagnostics.Append(diags...)
        if resp.Diagnostics.HasError() {
            return
        }
        scopesStr = strings.Join(scopes, " ")
    }

    // Extract the user email from the access token
    email, err := r.client.GetUserEmailFromToken()
    if err != nil {
        resp.Diagnostics.AddError(
            "Error extracting user email from access token",
            "Could not extract user email from access token, unexpected error: "+err.Error(),
        )
        return
    }

    // Merge OAuth values with extra configuration
    values, err := r.mergeCredentialValues(ctx, plan.OAuth, plan.ExtraConfiguration, scopesStr)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error merging credential values",
            "Could not merge OAuth and extra configuration values: "+err.Error(),
        )
        return
    }

    // Create the integration credentials
    createCredReq := client.CreateIntegrationCredentialsRequest{
        Name:          email,
        Values:        values,
        Provider:      integration.Type,
        Scheme:        "oauth_app", // Currently only oauth_app creds are supported
        IntegrationID: integrationID,
    }

    credential, err := r.client.CreateIntegrationCredentials(ctx, projectID, createCredReq)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating integration credentials",
            "Could not create integration credentials, unexpected error: "+err.Error(),
        )
        return
    }

    // Set the ID, scheme, and provider in the state
    plan.ID = types.StringValue(credential.ID)
    plan.Scheme = types.StringValue(credential.Scheme)
    plan.Provider = types.StringValue(credential.Provider)

    // Set state to fully populated data
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

// Read refreshes the Terraform state with the latest data.
func (r *integrationCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state integrationCredentialsResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := state.ProjectID.ValueString()
    credID := state.ID.ValueString()

    // Check if extra configuration was originally specified by the user
    hasExtraConfigInState := !state.ExtraConfiguration.IsNull() && !state.ExtraConfiguration.IsUnknown()
    
    // Add debug logging
    tflog.Debug(ctx, "Read - hasExtraConfigInState", map[string]interface{}{"hasExtraConfigInState": hasExtraConfigInState})

    // Retrieve the decrypted credential
    credential, err := r.client.GetDecryptedCredential(ctx, projectID, credID)
    if err != nil {
        if strings.Contains(err.Error(), "status code: 404") {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError(
            "Error retrieving decrypted credential",
            "Could not retrieve decrypted credential, unexpected error: "+err.Error(),
        )
        return
    }

    // Update the state with the retrieved data
    state.IntegrationID = types.StringValue(credential.IntegrationID)
    state.Scheme = types.StringValue(credential.Scheme)
    state.Provider = types.StringValue(credential.Provider)

    // Extract the client ID and secret from the decrypted credential values
    clientID, ok := credential.Values["clientId"].(string)
    if !ok {
        resp.Diagnostics.AddError(
            "Error extracting client ID",
            "Could not extract client ID from the decrypted credential values",
        )
        return
    }

    clientSecret, ok := credential.Values["clientSecret"].(string)
    if !ok {
        resp.Diagnostics.AddError(
            "Error extracting client secret",
            "Could not extract client secret from the decrypted credential values",
        )
        return
    }

    scopesStr, ok := credential.Values["scopes"].(string)
    var scopesList types.List
    if !ok {
        // If scopes are not found in the decrypted credential values, set them as null in the state
        scopesList = types.ListNull(types.StringType)
    } else {
        // If scopes are found and not empty, split them and store them in the state
        if scopesStr != "" {
            scopesArr := strings.Split(scopesStr, " ")
            var scopesAttr []attr.Value
            for _, scope := range scopesArr {
                scopesAttr = append(scopesAttr, types.StringValue(scope))
            }
            scopesList = types.ListValueMust(types.StringType, scopesAttr)
        } else {
            // If scopes are found but empty, set them as null in the state
            scopesList = types.ListNull(types.StringType)
        }
    }

    // Update the OAuth block in the state
    state.OAuth = &oauthModel{
        ClientID:     types.StringValue(clientID),
        ClientSecret: types.StringValue(clientSecret),
        Scopes:       scopesList,
    }

    // Only extract extra configuration if it was originally specified by the user
    if hasExtraConfigInState {
        // Extract all extra configuration keys from API response (excluding OAuth fields)
        filteredExtraConfig, err := r.extractAllExtraConfigurationFromAPI(ctx, credential.Values)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error extracting extra configuration",
                "Could not extract extra configuration from API response: "+err.Error(),
            )
            return
        }
        state.ExtraConfiguration = filteredExtraConfig
    } else {
        // If no extra configuration was specified, keep it as null
        state.ExtraConfiguration = types.DynamicNull()
    }

    // Set the refreshed state
    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *integrationCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan integrationCredentialsResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var state integrationCredentialsResourceModel
    diags = req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    scopesStr := ""

    // Handle provider type validation for extra configuration
    if state.Provider.ValueString() == "custom" {
        if !plan.OAuth.Scopes.IsNull() {
            resp.Diagnostics.AddError(
                "Invalid scopes",
                "Scopes must not be specified for custom integrations",
            )
            return
        }
    } else {
        if plan.OAuth.Scopes.IsNull() {
            resp.Diagnostics.AddError(
                "Missing scopes",
                "Scopes must be specified for non-custom integrations",
            )
            return
        }

        // Extract the scopes
        var scopes []string
        diags = plan.OAuth.Scopes.ElementsAs(ctx, &scopes, false)
        resp.Diagnostics.Append(diags...)
        if resp.Diagnostics.HasError() {
            return
        }
        scopesStr = strings.Join(scopes, " ")
    }

    projectID := plan.ProjectID.ValueString()
    integrationID := plan.IntegrationID.ValueString()
    credentialID := state.ID.ValueString()

    // Retrieve the integration for validation
    integration, err := r.client.GetIntegration(ctx, projectID, integrationID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error retrieving integration",
            "Could not retrieve integration, unexpected error: "+err.Error(),
        )
        return
    }

    // Validate extra configuration
    validationDiags := r.validateExtraConfiguration(ctx, plan.ExtraConfiguration, integration)
    resp.Diagnostics.Append(validationDiags...)
    if resp.Diagnostics.HasError() {
        return
    }

    _, err = r.client.GetDecryptedCredential(ctx, projectID, credentialID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error retrieving decrypted credential",
            "Could not retrieve decrypted credential, unexpected error: "+err.Error(),
        )
        return
    }

    // Extract the user email from the access token
    email, err := r.client.GetUserEmailFromToken()
    if err != nil {
        resp.Diagnostics.AddError(
            "Error extracting user email from access token",
            "Could not extract user email from access token, unexpected error: "+err.Error(),
        )
        return
    }

    // Merge OAuth and extra configuration values for update request
    // This preserves type conversion consistency with Create operation
    values, err := r.mergeCredentialValues(ctx, plan.OAuth, plan.ExtraConfiguration, scopesStr)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error merging credential values",
            "Could not merge OAuth and extra configuration values: "+err.Error(),
        )
        return
    }

    // Update the integration credentials
    updateCredReq := client.CreateIntegrationCredentialsRequest{
        Name:          email,
        Values:        values,
        Provider:      state.Provider.ValueString(),
        Scheme:        state.Scheme.ValueString(),
        IntegrationID: integrationID,
    }

    _, err = r.client.CreateIntegrationCredentials(ctx, projectID, updateCredReq)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error updating integration credentials",
            "Could not update integration credentials, unexpected error: "+err.Error(),
        )
        return
    }

    // Read the updated credential to ensure state consistency
    updatedDecryptedCredential, err := r.client.GetDecryptedCredential(ctx, projectID, credentialID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error retrieving updated decrypted credential",
            "Could not retrieve updated decrypted credential, unexpected error: "+err.Error(),
        )
        return
    }

    // Set the ID, scheme, and provider in the state
    plan.ID = types.StringValue(credentialID)
    plan.Scheme = types.StringValue(updatedDecryptedCredential.Scheme)
    plan.Provider = types.StringValue(updatedDecryptedCredential.Provider)

    // Extract and set the OAuth values from the updated credential
    clientID, ok := updatedDecryptedCredential.Values["clientId"].(string)
    if !ok {
        resp.Diagnostics.AddError(
            "Error extracting client ID from updated credential",
            "Could not extract client ID from the updated decrypted credential values",
        )
        return
    }

    clientSecret, ok := updatedDecryptedCredential.Values["clientSecret"].(string)
    if !ok {
        resp.Diagnostics.AddError(
            "Error extracting client secret from updated credential",
            "Could not extract client secret from the updated decrypted credential values",
        )
        return
    }

    updatedScopesStr, ok := updatedDecryptedCredential.Values["scopes"].(string)
    var scopesList types.List
    if !ok {
        scopesList = types.ListNull(types.StringType)
    } else {
        if updatedScopesStr != "" {
            scopesArr := strings.Split(updatedScopesStr, " ")
            var scopesAttr []attr.Value
            for _, scope := range scopesArr {
                scopesAttr = append(scopesAttr, types.StringValue(scope))
            }
            scopesList = types.ListValueMust(types.StringType, scopesAttr)
        } else {
            scopesList = types.ListNull(types.StringType)
        }
    }

    // Update the OAuth block in the state
    plan.OAuth = &oauthModel{
        ClientID:     types.StringValue(clientID),
        ClientSecret: types.StringValue(clientSecret),
        Scopes:       scopesList,
    }

    // Check if extra configuration was originally specified by the user
    hasExtraConfigInPlan := !plan.ExtraConfiguration.IsNull() && !plan.ExtraConfiguration.IsUnknown()
    
    // Add debug logging
    tflog.Debug(ctx, "Update - hasExtraConfigInPlan", map[string]interface{}{"hasExtraConfigInPlan": hasExtraConfigInPlan})
    tflog.Debug(ctx, "Update - API response values", map[string]interface{}{"values": updatedDecryptedCredential.Values})
    
    // Only extract extra configuration if it was originally specified by the user
    if hasExtraConfigInPlan {
        updatedExtraConfig, err := r.extractAllExtraConfigurationFromAPI(ctx, updatedDecryptedCredential.Values)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error extracting extra configuration from updated credential",
                "Could not extract extra configuration from updated credential: "+err.Error(),
            )
            return
        }
        tflog.Debug(ctx, "Update - extracted extra config", map[string]interface{}{"updatedExtraConfig": updatedExtraConfig})
        plan.ExtraConfiguration = updatedExtraConfig
    } else {
        // If no extra configuration was specified, keep it as null
        tflog.Debug(ctx, "Update - setting extra config to null")
        plan.ExtraConfiguration = types.DynamicNull()
    }

    // Set state to fully populated data
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *integrationCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state integrationCredentialsResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := state.ProjectID.ValueString()
    credentialID := state.ID.ValueString()

    err := r.client.DeleteCredentials(ctx, projectID, credentialID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error deleting credentials",
            "Could not delete credentials, unexpected error: "+err.Error(),
        )
        return
    }
}