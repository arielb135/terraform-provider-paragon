---
page_title: "integration_status Resource - paragon"
subcategory: ""
description: |-
  Activates/deactivates an integration.
---

# paragon_integration_status (Resource)

Controls the enablement of an integration.

## Example Usage

Use `paragon_integrations` data source to find out the relevant `integration_id`.

```terraform
resource "paragon_integration_status" "example" {
  integration_id = "f6ab5c54-fc30-4232-973d-73486ca708fc"
  project_id = "69b05bc7-4996-4b4e-888b-3a67915ee1d8"
  active = true
}
```

## Schema

### Argument Reference

- `integration_id` (String, Required) Identifier of the integration to enable.
- `project_id` (String, Required) Identifier of the project of the integration to enable.
- `active` (Object, Required) weather the integration is active or not.

### Attributes Reference

- `id` (String) Identifier of the integration. Same as `integration_id`.

## JSON State Structure Example

Here's a state sample

```json
{
    "active": true,
    "id": "f6ab5c54-fc30-4232-973d-73486ca708fc",
    "integration_id": "f6ab5c54-fc30-4232-973d-73486ca708fc",
    "project_id": "69b05bc7-4996-4b4e-888b-3a67915ee1d8"
}
```
