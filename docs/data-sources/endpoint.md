---
page_title: "monitrs_endpoint Data Source - monitrs"
subcategory: ""
description: |-
  Look up a monitored endpoint by ID.
---

# monitrs_endpoint (Data Source)

Look up a monitored endpoint by its UUID. Useful for referencing endpoints
created outside Terraform.

## Example Usage

```terraform
data "monitrs_endpoint" "existing" {
  id = "7f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c"
}
```

## Schema

### Required

- `id` (String) The endpoint UUID.

### Read-Only

- `name` (String)
- `url` (String)
- `method` (String)
- `interval_seconds` (Number)
- `is_active` (Boolean)
- `openapi_spec_url` (String)
- `heartbeat_token` (String, Sensitive)
- `created_at` (String)
