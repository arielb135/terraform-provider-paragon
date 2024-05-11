package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/arielb135/terraform-provider-paragon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &paragonProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &paragonProvider{
			version: version,
		}
	}
}

// paragonProvider is the provider implementation.
type paragonProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// paragonProviderModel maps provider schema data to a Go type.
type paragonProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	BaseURL  types.String `tfsdk:"base_url"`
}

// Metadata returns the provider type name.
func (p *paragonProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "paragon"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *paragonProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username for authenticating with the Paragon service.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password for authenticating with the Paragon service.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "The base URL of the Paragon service. Defaults to 'https://zeus.useparagon.com'.",
			},
		},
	}
}

// Configure prepares a Paragon API client for data sources and resources.
//
//gocyclo:ignore
func (p *paragonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Paragon client")

	// Retrieve provider data from configuration
	var config paragonProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Paragon Username",
			"The provider cannot create the Paragon API client as there is an unknown configuration value for the Paragon API username.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Paragon Password",
			"The provider cannot create the Paragon API client as there is an unknown configuration value for the Paragon API password.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Set the base URL, using the default value if not provided
	baseURL := "https://zeus.useparagon.com"
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	// Create the Paragon API client
	api := client.NewClient(baseURL)

    // Authenticate with the Paragon service
    err := api.Authenticate(ctx, config.Username.ValueString(), config.Password.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Authenticate with Paragon API",
            "An unexpected error occurred when authenticating with the Paragon API. "+
                "If the error is not clear, please contact the provider developers.\n\n"+
                "Paragon Client Error: "+err.Error(),
        )
        return
    }

    // Make the Paragon client available during DataSource and Resource
    // type Configure methods.
    resp.DataSourceData = api
    resp.ResourceData = api

	tflog.Info(ctx, "Configured Paragon client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *paragonProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewOrganizationsDataSource,
        NewOrganizationDataSource,
        NewTeamsDataSource,
        NewTeamDataSource,
        NewIntegrationsDataSource,
        NewWorkflowDataSource,
        NewWorkflowsDataSource,
    }
}

// Resources defines the resources implemented in the provider.
func (p *paragonProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewProjectResource,
        NewSDKKeysResource,
        NewEnvironmentSecretResource,
        NewTeamMemberResource,
        NewCLIKeyResource,
        NewIntegrationCredentialsResource,
        NewIntegrationStatusResource,
        NewEventsDestinationResource,
        NewWorkflowDeploymentResource,
    }
}