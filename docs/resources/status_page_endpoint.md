---
page_title: "monitrs_status_page_endpoint Resource - monitrs"
subcategory: ""
description: |-
  Attaches a monitored endpoint to a status page.
---

# monitrs_status_page_endpoint (Resource)

Attaches a `monitrs_endpoint` to a `monitrs_status_page`. Each attachment
carries its own `display_name` and (optional) `display_order` for the page's
visitors — this can differ from the endpoint's internal `name`.

All fields are `RequiresReplace` because the underlying API supports only
POST + DELETE on attachments (no PATCH). Any change recreates the attachment.

## Example Usage

Included in the [`monitrs_status_page`](status_page) example.

## Schema

### Required

- `page_id` (String) Status page UUID (`monitrs_status_page.id`).
- `endpoint_id` (String) Endpoint UUID (`monitrs_endpoint.id`).
- `display_name` (String) Public-facing name shown on the status page (can differ from the endpoint's internal name).

### Optional

- `display_order` (Number) Sort order on the page.

### Read-Only

- `id` (String) Attachment UUID.
