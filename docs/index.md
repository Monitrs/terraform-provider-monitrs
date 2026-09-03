---
page_title: "monitrs Provider"
subcategory: ""
description: |-
  Terraform provider for monit.rs — schema-aware API monitoring.
---

# monitrs Provider

Manage monit.rs endpoints, status pages, and alert channels via Terraform.
For the full API reference, see [monit.rs/api-docs](https://monit.rs/api-docs).

Get an API key at [monit.rs/settings](https://monit.rs/settings) → API Keys.
Scope it with least privilege — the provider only needs `write:endpoints`,
`write:status-pages`, and `write:alerts` for a full IaC workflow.

## Example Usage

```terraform
{{tffile "examples/provider/provider.tf"}}
```

## Schema

### Optional

- `api_key` (String, Sensitive) monit.rs API key (starts with `mrs_live_`). Falls back to the `MONITRS_API_KEY` env var if unset.
- `base_url` (String) Override the API base URL. Defaults to `https://monit.rs`. Set for self-hosted or staging deploys.
