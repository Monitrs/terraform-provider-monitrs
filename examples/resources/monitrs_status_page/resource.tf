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
