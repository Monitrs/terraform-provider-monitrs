terraform {
  required_providers {
    monitrs = {
      source  = "Monitrs/monitrs"
      version = "~> 0.1"
    }
  }
}

provider "monitrs" {
  # api_key = "mrs_live_..." — or set MONITRS_API_KEY
}
