package provider

import (
    "context"
    "fmt"
    "time"
    "strings"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/arielb135/terraform-provider-paragon/internal/client"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

)

var (
    _ resource.Resource              = &workflowDeploymentResource{}
    _ resource.ResourceWithConfigure = &workflowDeploymentResource{}
)

func NewWorkflowDeploymentResource() resource.Resource {
    return &workflowDeploymentResource{}
}

type workflowDeploymentResource struct {
    client *client.Client
}

type workflowDeploymentResourceModel struct {
    ID         types.String `tfsdk:"id"`
    ProjectID  types.String `tfsdk:"project_id"`
    WorkflowID types.String `tfsdk:"workflow_id"`
    Version    types.Int64  `tfsdk:"version"`
    Deployed   types.Bool   `tfsdk:"deployed"`
}

func (r *workflowDeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*client.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.client = client
}

func (r *workflowDeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_workflow_deployment"
}

func (r *workflowDeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manages a workflow deployment in Paragon.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "The ID of the workflow deployment.",
                Computed:    true,
            },
            "project_id": schema.StringAttribute{
                Description: "The ID of the project.",
                Required:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "workflow_id": schema.StringAttribute{
                Description: "The ID of the workflow.",
                Required:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "version": schema.Int64Attribute{
                Description: "The version of the workflow deployment. Change this to trigger a new deployment",
                Required:    true,
            },
            "deployed": schema.BoolAttribute{
                Description: "Indicates if the workflow is deployed.",
                Computed:    true,
            },
        },
    }
}

func (r *workflowDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan workflowDeploymentResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := plan.ProjectID.ValueString()
    workflowID := plan.WorkflowID.ValueString()

    deploymentID, err := r.client.CreateWorkflowDeployment(ctx, projectID, workflowID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating workflow deployment",
            "Could not create workflow deployment, unexpected error: "+err.Error(),
        )
        return
    }

    for {
        deployment, err := r.client.GetWorkflowDeployment(ctx, projectID, deploymentID)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error retrieving workflow deployment status",
                "Could not retrieve workflow deployment status, unexpected error: "+err.Error(),
            )
            return
        }

        if deployment.Status != "DEPLOYING" {
            if deployment.Status == "DEPLOYED" {
                plan.ID = types.StringValue(deploymentID)
                plan.Deployed = types.BoolValue(true)
                break
            } else {
                resp.Diagnostics.AddError(
                    "Error creating workflow deployment",
                    fmt.Sprintf("Workflow deployment failed with status: %s", deployment.Status),
                )
                return
            }
        }

        time.Sleep(2 * time.Second)
    }

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *workflowDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state workflowDeploymentResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := state.ProjectID.ValueString()
    workflowID := state.WorkflowID.ValueString()

    migration, err := r.client.GetLatestWorkflowMigration(ctx, projectID, workflowID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error reading workflow deployment",
            "Could not read workflow deployment, unexpected error: "+err.Error(),
        )
        return
    }

    if migration == nil {
        resp.State.RemoveResource(ctx)
        return
    }

    state.ID = types.StringValue(migration.Deployment.ID)

    // Check if the deployed flag has changed
    if state.Deployed.ValueBool() != migration.Deployment.IsActive {
        // If the deployed flag has changed, recreate the resource
        resp.Diagnostics.AddWarning(
            "Workflow was undeployed outside terraform, Removing from state...",
            "The workflow deployment state has changed outside of Terraform.",
        )
        resp.State.RemoveResource(ctx)
        return
    }

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
}

func (r *workflowDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan workflowDeploymentResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := plan.ProjectID.ValueString()
    workflowID := plan.WorkflowID.ValueString()

    deploymentID, err := r.client.CreateWorkflowDeployment(ctx, projectID, workflowID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error updating workflow deployment",
            "Could not update workflow deployment, unexpected error: "+err.Error(),
        )
        return
    }

    for {
        deployment, err := r.client.GetWorkflowDeployment(ctx, projectID, deploymentID)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error retrieving workflow deployment status",
                "Could not retrieve workflow deployment status, unexpected error: "+err.Error(),
            )
            return
        }

        if deployment.Status != "DEPLOYING" {
            if deployment.Status == "DEPLOYED" {
                plan.ID = types.StringValue(deploymentID)
                plan.Deployed = types.BoolValue(true)
                break
            } else {
                resp.Diagnostics.AddError(
                    "Error updating workflow deployment",
                    fmt.Sprintf("Workflow deployment failed with status: %s", deployment.Status),
                )
                return
            }
        }

        time.Sleep(2 * time.Second)
    }

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *workflowDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state workflowDeploymentResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectID := state.ProjectID.ValueString()
    workflowID := state.WorkflowID.ValueString()

    // Call DeleteWorkflowDeployment to initiate the undeployment
    err := r.client.DeleteWorkflowDeployment(ctx, projectID, workflowID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error deleting workflow deployment",
            "Could not delete workflow deployment, unexpected error: "+err.Error(),
        )
        return
    }

    for {
        deployment, err := r.client.GetWorkflowDeployment(ctx, projectID, state.ID.ValueString())
        if err != nil {
            if strings.Contains(err.Error(), "status code: 404") {
                // The deployment is deleted (not found), break the loop
                break
            }
            resp.Diagnostics.AddError(
                "Error retrieving workflow deployment status",
                "Could not retrieve workflow deployment status, unexpected error: "+err.Error(),
            )
            return
        }

        if deployment.Status == "UNDEPLOYED" {
            // The deployment is successfully undeployed, break the loop
            break
        } else if deployment.Status != "UNDEPLOYING" {
            resp.Diagnostics.AddError(
                "Error deleting workflow deployment",
                fmt.Sprintf("Workflow deployment failed to undeploy with status: %s", deployment.Status),
            )
            return
        }

        time.Sleep(2 * time.Second)
    }
}