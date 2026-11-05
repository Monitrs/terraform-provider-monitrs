---
page_title: "monitrs_status_page Resource - monitrs"
subcategory: ""
description: |-
  A public or private status page hosted at <slug>.status.monit.rs.
---

# monitrs_status_page (Resource)

A status page hosted at `<slug>.status.monit.rs`. Attach endpoints to it via
the [`monitrs_status_page_endpoint`](status_page_endpoint) resource.

## Example Usage

```terraform
resource "monitrs_status_page" "prod" {
  slug        = "prod"
  title       = "Example Corp API status"
  description = "Real-time status of our public API surfaces."
  is_public   = true
}

resource "monitrs_status_page_endpoint" "prod_api_health" {
  page_id       = monitrs_status_page.prod.id
  endpoint_id   = monitrs_endpoint.api_health.id
  display_name  = "API health check"
  display_order = 1
}
```

## Schema

### Required

- `slug` (String) URL slug. Rendered as `<slug>.status.monit.rs`. Cannot be changed after creation (recreates on change).
- `title` (String) Page title shown to visitors.

### Optional

- `description` (String) Longer-form description.
- `is_public` (Boolean) Whether the page is publicly reachable. Default `true`.

### Read-Only

- `id` (String) monit.rs status page UUID.
- `created_at` (String) ISO-8601 timestamp.

## Import

```shell
terraform import monitrs_status_page.example 7f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c
```
