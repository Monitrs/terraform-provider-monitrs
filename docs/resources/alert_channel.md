---
page_title: "monitrs_alert_channel Resource - monitrs"
subcategory: ""
description: |-
  Where monit.rs sends alerts when incidents open.
---

# monitrs_alert_channel (Resource)

Where monit.rs sends alerts on incident open / regression detected.

All fields are `RequiresReplace` because the underlying API supports only
POST + DELETE on channels (no PATCH). Any change recreates the channel.

## Example Usage

```terraform
resource "monitrs_alert_channel" "slack_prod" {
  channel_type = "slack"
  config = {
    url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  }
  min_severity = "error"
}
```

## Schema

### Required

- `channel_type` (String) One of `email`, `slack`, `discord`, `teams`, `pagerduty`, `webhook`.
- `config` (Map of String, Sensitive) Channel-specific config. For URL-based channels (`slack`, `discord`, `webhook`), typically `{ url = "https://..." }`. For `email`, typically `{ address = "..." }`. For `pagerduty`, typically `{ integration_key = "..." }`. Only string values are supported in v0.1.

### Optional

- `min_severity` (String) Only fire this channel for incidents at or above this severity. One of `critical`, `error`, `warning`, `info`.

### Read-Only

- `id` (String) monit.rs alert channel UUID.
- `created_at` (String) ISO-8601 timestamp.
