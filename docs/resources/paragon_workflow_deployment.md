---
page_title: "workflow_deployment Resource - paragon"
subcategory: ""
description: |-
  Manages the deployment of a workflow.
---

# paragon_workflow_deployment (Resource)

Controls the deployment of a workflow within a project - Use this resource to enable a workflow.

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
}
```

## Schema

### Argument Reference
- `project_id` (String, Required) Identifier of the project of the integration to enable.
- `workflow_id` (String, Required) Identifier of the workflow to deploy.

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
  "deployed": true
}
```