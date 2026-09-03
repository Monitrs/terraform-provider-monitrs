resource "monitrs_endpoint" "api_health" {
  name             = "prod api /health"
  url              = "https://api.example.com/health"
  method           = "GET"
  interval_seconds = 300
}
