# terraform-provider-monitrs

<p>
  <a href="https://registry.terraform.io/providers/Monitrs/monitrs/latest"><img alt="Terraform Registry" src="https://img.shields.io/badge/Terraform%20Registry-Monitrs%2Fmonitrs-7B42BC?logo=terraform"></a>
  <a href="https://github.com/Monitrs/terraform-provider-monitrs/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/Monitrs/terraform-provider-monitrs?display_name=tag&sort=semver"></a>
  <a href="./LICENSE"><img alt="MPL-2.0 License" src="https://img.shields.io/badge/license-MPL--2.0-blue"></a>
</p>

Official Terraform provider for [monit.rs](https://monit.rs) — declare your
**API monitors, alert routing, and public status pages as code**. Full
reference on the
[Terraform Registry](https://registry.terraform.io/providers/Monitrs/monitrs/latest/docs).

- 📡 **Monitors as code** — endpoints, probe intervals, and OpenAPI contract validation
- 🔔 **Alert routing** — email, Slack, Discord, Teams, PagerDuty, or a raw webhook, with per-channel severity floors
- 📊 **Status pages** — a public page at `<slug>.status.monit.rs`, with the endpoints you choose to expose
- 🔑 **Scoped API keys** — the provider authenticates with a scoped key, not an account password

## Install

```hcl
terraform {
  required_providers {
    monitrs = {
      source  = "Monitrs/monitrs"
      version = "~> 0.1"
    }
  }
}

provider "monitrs" {
  # Prefers the MONITRS_API_KEY environment variable.
  # api_key = "mrs_live_..."
}
```

Create a key under **Settings → API keys** at [monit.rs](https://monit.rs). Keep
it out of state files and out of version control — export it instead:

```sh
export MONITRS_API_KEY="mrs_live_..."
```

## A complete example

Monitor an API, route failures to Slack, and publish a status page:

```hcl
resource "monitrs_endpoint" "checkout" {
  name             = "Checkout API"
  url              = "https://api.example.com/v1/health"
  method           = "GET"
  interval_seconds = 300

  # Optional: validate every response against your spec, so a contract
  # break is caught even while the endpoint is still returning 200.
  openapi_spec_url = "https://api.example.com/openapi.json"
}

resource "monitrs_alert_channel" "oncall" {
  channel_type = "slack"
  min_severity = "error"

  config = {
    url = var.slack_webhook_url
  }
}

resource "monitrs_status_page" "public" {
  slug        = "example"
  title       = "Example API status"
  description = "Live availability for the Example public API."
  is_public   = true
}

resource "monitrs_status_page_endpoint" "checkout" {
  page_id       = monitrs_status_page.public.id
  endpoint_id   = monitrs_endpoint.checkout.id
  display_name  = "Checkout"
  display_order = 1
}

output "status_page_url" {
  value = "https://${monitrs_status_page.public.slug}.status.monit.rs"
}
```

## Resources and data sources

| Resource | Purpose |
| --- | --- |
| `monitrs_endpoint` | An API endpoint to probe, with interval and optional OpenAPI contract validation |
| `monitrs_alert_channel` | Where incidents are delivered, with an optional minimum severity |
| `monitrs_status_page` | A public status page served at `<slug>.status.monit.rs` |
| `monitrs_status_page_endpoint` | Attaches an endpoint to a status page under a public display name |

| Data source | Purpose |
| --- | --- |
| `monitrs_me` | The authenticated account: tier, auth kind, and the scopes the key carries |
| `monitrs_endpoint` | Look up a single existing endpoint |
| `monitrs_endpoints` | List existing endpoints, to reference monitors Terraform does not own |

`monitrs_endpoint.heartbeat_token` is exported for cron-liveness checks — ping
`POST https://monit.rs/api/v1/heartbeats/{token}` from a scheduled job and the
monitor fails if the ping stops arriving. It is marked sensitive, so it will not
appear in plan output.

## Probe interval minimums

`interval_seconds` is validated server-side against your plan. Setting a value
below your tier's floor is rejected at apply time rather than silently rounded:

| Tier | Minimum |
| ---: | ---: |
| Free | 900s |
| Developer | 300s |
| Starter | 60s |
| Pro / Scale | 30s |

## Status

**v0.1.0 — pre-1.0.** The resource shapes above are what the provider ships
today, but minor releases may still change defaults. Pin with
`version = "~> 0.1"` and read the release notes before upgrading.

`config` on `monitrs_alert_channel` accepts string values only in v0.1.

## Related

- [monit.rs OpenAPI diff — GitHub Action](https://github.com/marketplace/actions/monit-rs-openapi-diff) — the same diff engine on every pull request
- [Public API changelog](https://monit.rs/changelog) — breaking changes across 47 tracked vendor APIs, no signup
- [API reference](https://monit.rs/api-docs) — the HTTP API this provider is built on

## Development

See [`PUBLISHING.md`](./PUBLISHING.md) for the release process. Issues and pull
requests are welcome at
[Monitrs/terraform-provider-monitrs](https://github.com/Monitrs/terraform-provider-monitrs).

## License

[MPL-2.0](./LICENSE)
