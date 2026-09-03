---
page_title: "monitrs_endpoints Data Source - monitrs"
subcategory: ""
description: |-
  List all monitored endpoints on the authenticated account.
---

# monitrs_endpoints (Data Source)

Lists monitored endpoints on the authenticated account. Useful for
enumerating existing infrastructure or building dashboards.

## Example Usage

```terraform
data "monitrs_endpoints" "all" {}

output "endpoint_names" {
  value = [for e in data.monitrs_endpoints.all.endpoints : e.name]
}
```

## Schema

### Optional

- `limit` (Number) Max endpoints to return. Default 50, max 200.

### Read-Only

- `endpoints` (List of Object) Every endpoint on the account. Fields: `id`, `name`, `url`, `method`, `interval_seconds`, `is_active`, `openapi_spec_url`, `heartbeat_token`, `created_at`.
