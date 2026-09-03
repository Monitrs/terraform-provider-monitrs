resource "monitrs_alert_channel" "slack_prod" {
  channel_type = "slack"
  config = {
    url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  }
  min_severity = "error"
}
