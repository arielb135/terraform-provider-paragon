---
page_title: "paragon_workflow Data Source - paragon"
subcategory: ""
description: |-
  Fetches a specific workflow by description.
---

# paragon_workflow (Data Source)

Fetches a specific workflow by description. Currently only ID is returned.

## Example Usage

```terraform
# Read a specific workflow by project ID, integration ID, and description
data "paragon_workflow" "example" {
  project_id     = "your_project_id"
  integration_id = "your_integration_id"
  description    = "the name you've given to your workflow"
}
```

## Errors
Error will be thrown if the workflow is not found.

## Schema

### Argument Reference

- `project_id` (String, Required) The ID of the project.
- `integration_id` (String, Required) The ID of the integration.
- `description` (String, Required) The description of the workflow to search for.

### Attributes Reference

- `organization` (Attributes) The organization details.

The `organization` block contains:

- `id` (String)  The ID of the workflow.

## JSON State Structure Example

Here's a state sample:

```json
{
  "project_id": "<project_id>",
  "integration_id": "<integration_id>",
  "description": "Create tickets from issues",
  "id": "<workflow_id>"
}
```
