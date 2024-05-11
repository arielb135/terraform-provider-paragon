---
page_title: "workflow_deployment Resource - paragon"
subcategory: ""
description: |-
  Manages the deployment of a workflow.
---

# paragon_workflow_deployment (Resource)

Controls the deployment of a workflow within a project - Use this resource to enable a workflow.

~> **IMPORTANT:** 
The integration must be enabled in order for the workflow to be deployed. If you manage the integration via terraform, make sure to add depends_on on the `paragon_integration_status` resource to avoid race conditions. 

### Depends on example
```terraform
data "paragon_workflow" "wfdata" {
  project_id     = "e0da0789-cd90-4ca7-897b-8a89404eb329"
  integration_id = "fb549b70-658b-4a14-9318-4dca3a88bfa7"
  description    = "your workflow description"
}

resource "paragon_integration_status" "integration_status" {
  project_id = "e0da0789-cd90-4ca7-897b-8a89404eb329"
  integration_id = "fb549b70-658b-4a14-9318-4dca3a88bfa7"
  active = true
}

resource "paragon_workflow_deployment" "enable_this_workflow" {
  project_id  = "e0da0789-cd90-4ca7-897b-8a89404eb329"
  workflow_id = data.paragon_workflow.wfdata.id
  version     = 1
  depends_on  = [paragon_integration_status.integration_status]
}
```


## Example Usage

Use `paragon_workflow` data source to find out the relevant `workflow_id`.


```terraform
data "paragon_workflow" "wfdata" {
  project_id     = "e0da0789-cd90-4ca7-897b-8a89404eb329"
  integration_id = "fb549b70-658b-4a14-9318-4dca3a88bfa7"
  description    = "your workflow description"
}

resource "paragon_workflow_deployment" "enable_this_workflow" {
  project_id  = "e0da0789-cd90-4ca7-897b-8a89404eb329"
  workflow_id = data.paragon_workflow.wfdata.id
  version     = 1
}
```

## Schema

### Argument Reference
- `project_id` (String, Required) Identifier of the project of the integration to enable.
- `workflow_id` (String, Required) Identifier of the workflow to deploy.
- `version` (Number, Required) The sole purpose of the version is to trigger latest deployment - if you had changes in your deployment, and you wish to deploy them - change the version number to any number different from what you have in the state.

### Attributes Reference

- `id` (String) Identifier of the workflow deployment, Used to track the latest deployment.
- `deployed` (Boolean) Indicates whether the workflow is deployed or not - Should be `true`.

## JSON State Structure Example

Here's a state sample

```json
{
  "id": "c4a133f8-0314-463f-b3f1-d566bb69a7c1",
  "project_id": "e0da0789-cd90-4ca7-897b-8a89404eb329",
  "workflow_id": "fbf8d8b8-c1e7-57a1-b7a8-6a8f895e940f",
  "version": 1,
  "deployed": true
}
```
