package provider

import (
    "context"
    "strings"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/arielb135/terraform-provider-paragon/internal/client"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"

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
    ID            types.String `tfsdk:"id"`
    ProjectID     types.String `tfsdk:"project_id"`
    IntegrationID types.String `tfsdk:"integration_id"`
    Scheme        types.String `tfsdk:"scheme"`
    Provider      types.String `tfsdk:"creds_provider"`
    OAuth         *oauthModel  `tfsdk:"oauth"`
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
        },
    }
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

    // Create the integration credentials
    createCredReq := client.CreateIntegrationCredentialsRequest{
        Name: email,
        Values: client.OAuthValues{
            ClientID:     plan.OAuth.ClientID.ValueString(),
            ClientSecret: plan.OAuth.ClientSecret.ValueString(),
            Scopes:       scopesStr,
        },
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
    if !ok {
        // If scopes are not found in the decrypted credential values, set them as null in the state
        state.OAuth.Scopes = types.ListNull(types.StringType)
    } else {
        // If scopes are found and not empty, split them and store them in the state
        if scopesStr != "" {
            scopesArr := strings.Split(scopesStr, " ")
            var scopesAttr []attr.Value
            for _, scope := range scopesArr {
                scopesAttr = append(scopesAttr, types.StringValue(scope))
            }
            state.OAuth.Scopes = types.ListValueMust(types.StringType, scopesAttr)
        } else {
            // If scopes are found but empty, set them as null in the state
            state.OAuth.Scopes = types.ListNull(types.StringType)
        }
    }

    // Update the OAuth block in the state
    state.OAuth = &oauthModel{
        ClientID:     types.StringValue(clientID),
        ClientSecret: types.StringValue(clientSecret),
        Scopes:       state.OAuth.Scopes,
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
    }


    projectID := plan.ProjectID.ValueString()
    integrationID := plan.IntegrationID.ValueString()
    credentialID := state.ID.ValueString()

    _, err := r.client.GetDecryptedCredential(ctx, projectID, credentialID)
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

    // Extract the scopes
    var scopes []string
    diags = plan.OAuth.Scopes.ElementsAs(ctx, &scopes, false)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
    scopesStr := strings.Join(scopes, " ")

    // Update the integration credentials
    updateCredReq := client.CreateIntegrationCredentialsRequest{
        Name:          email,
        Values:        client.OAuthValues{
            ClientID:     plan.OAuth.ClientID.ValueString(),
            ClientSecret: plan.OAuth.ClientSecret.ValueString(),
            Scopes:       scopesStr,
        },
        Provider:      state.Provider.ValueString(),
        Scheme:        state.Scheme.ValueString(),
        IntegrationID: integrationID,
    }

    updatedCredential, err := r.client.CreateIntegrationCredentials(ctx, projectID, updateCredReq)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error updating integration credentials",
            "Could not update integration credentials, unexpected error: "+err.Error(),
        )
        return
    }

    // Set the ID, scheme, and provider in the state
    plan.ID = types.StringValue(credentialID)
    plan.Scheme = types.StringValue(updatedCredential.Scheme)
    plan.Provider = types.StringValue(updatedCredential.Provider)

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