package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/arielb135/terraform-provider-paragon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
    _ datasource.DataSource              = &integrationsDataSource{}
    _ datasource.DataSourceWithConfigure = &integrationsDataSource{}
)

// NewIntegrationsDataSource is a helper function to simplify the provider implementation.
func NewIntegrationsDataSource() datasource.DataSource {
    return &integrationsDataSource{}
}

// integrationsDataSource is the data source implementation.
type integrationsDataSource struct {
    client *client.Client
}

// integrationsDataSourceModel maps the data source schema data.
type integrationsDataSourceModel struct {
    ProjectID    types.String                     `tfsdk:"project_id"`
    Integrations map[string]integrationModel      `tfsdk:"integrations"`
}

type integrationModel struct {
    ID                  types.String `tfsdk:"id"`
    CustomIntegrationID types.String `tfsdk:"custom_integration_id"`
    AuthenticationType types.String  `tfsdk:"authentication_type"`
    Type                types.String `tfsdk:"type"`
    IsActive            types.Bool   `tfsdk:"is_active"`
    ConnectedUserCount  types.Int64  `tfsdk:"connected_user_count"`
}

// Configure adds the provider configured client to the data source.
func (d *integrationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*client.Client)
    if !ok {
        return
    }
    d.client = client
}

// Metadata returns the data source type name.
func (d *integrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_integrations"
}

// Schema defines the schema for the data source.
func (d *integrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Fetches the list of integrations for a project.",
        Attributes: map[string]schema.Attribute{
            "project_id": schema.StringAttribute{
                Description: "The ID of the project.",
                Required:    true,
            },
            "integrations": schema.MapNestedAttribute{
                Description: "The map of integrations keyed by their type.",
                Computed:    true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "id": schema.StringAttribute{
                            Description: "The ID of the integration.",
                            Computed:    true,
                        },
                        "custom_integration_id": schema.StringAttribute{
                            Description: "The custom integration ID in case of a custom integration.",
                            Computed:    true,
                        },
                        "authentication_type": schema.StringAttribute{
                            Description: "In case of a custom integration, the authentication type.",
                            Computed:    true,
                        },
                        "type": schema.StringAttribute{
                            Description: "The type of the integration. If the type is 'custom', it will contain the slug of the custom integration.",
                            Computed:    true,
                        },
                        "is_active": schema.BoolAttribute{
                            Description: "Indicates if the integration is active.",
                            Computed:    true,
                        },
                        "connected_user_count": schema.Int64Attribute{
                            Description: "The count of connected users for the integration.",
                            Computed:    true,
                        },
                    },
                },
            },
        },
    }
}

// Read refreshes the Terraform state with the latest data.
func (d *integrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var state integrationsDataSourceModel

    // Get the project ID from the configuration.
    diags := req.Config.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := state.ProjectID.ValueString()

    // Retrieve the integrations using the GetIntegrations function.
    integrations, err := d.client.GetIntegrations(ctx, projectID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Integrations",
            err.Error(),
        )
        return
    }

    // Map the integrations to the state.
    integrationModels := make(map[string]integrationModel)
    for _, integration := range integrations {
        integrationType := integration.Type
        authenticationType := ""
        if integration.Type == "custom" && integration.CustomIntegration != nil {
            integrationType = integration.CustomIntegration.Slug
            authenticationType = integration.CustomIntegration.AuthenticationType
        }

        var customIntegrationID types.String
        if integration.CustomIntegrationID == nil {
            customIntegrationID = types.StringNull()
        } else {
            customIntegrationID = types.StringValue(*integration.CustomIntegrationID)
        }

        integrationModels[integrationType] = integrationModel{
            ID:                  types.StringValue(integration.ID),
            CustomIntegrationID: customIntegrationID,
            Type:                types.StringValue(integrationType),
            IsActive:            types.BoolValue(integration.IsActive),
            AuthenticationType: types.StringValue(authenticationType),
            ConnectedUserCount:  types.Int64Value(int64(integration.ConnectedUserCount)),
        }
    }

    state.Integrations = integrationModels

    // Set the state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}