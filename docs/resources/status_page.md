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
{{tffile "examples/resources/monitrs_status_page/resource.tf"}}
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
