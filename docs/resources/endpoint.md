---
page_title: "monitrs_endpoint Resource - monitrs"
subcategory: ""
description: |-
  A monitored endpoint. monit.rs probes this on the configured interval,
  learns its response schema, and opens regressions on schema drift.
---

# monitrs_endpoint (Resource)

Represents a single HTTPS endpoint that monit.rs probes on the configured
interval. monit.rs learns the response schema, opens regressions on drift,
and (with an `openapi_spec_url` attached) runs contract validation on every
probe.

## Example Usage

```terraform
resource "monitrs_endpoint" "api_health" {
  name             = "prod api /health"
  url              = "https://api.example.com/health"
  method           = "GET"
  interval_seconds = 300
}
```

## Schema

### Required

- `name` (String) Human-readable name.
- `url` (String) HTTPS URL to probe.

### Optional

- `method` (String) HTTP method. Default `GET`.
- `interval_seconds` (Number) Probe interval in seconds. Minimum depends on tier: Free 900, Developer 300, Starter 60, Pro/Scale 30. Default 900.
- `openapi_spec_url` (String) URL of the OpenAPI spec for this endpoint. When set, enables contract validation on every probe.
- `is_active` (Boolean) Whether the endpoint is actively probed. Default `true`.

### Read-Only

- `id` (String) monit.rs endpoint UUID.
- `heartbeat_token` (String, Sensitive) Opaque token for cron-liveness heartbeat pings. Use in `POST https://monit.rs/api/v1/heartbeats/{token}`.
- `created_at` (String) ISO-8601 timestamp.

## Import

Endpoints can be imported by their UUID:

```shell
terraform import monitrs_endpoint.example 7f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c
```
