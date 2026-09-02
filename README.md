# terraform-provider-monitrs

Official Terraform provider for [monit.rs](https://monit.rs). IaC your
endpoints, status pages, and alert channels. Full docs on the
[Terraform Registry](https://registry.terraform.io/providers/Monitrs/monitrs/latest/docs).

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
  # api_key = "mrs_live_..." (or MONITRS_API_KEY env var)
}
```

**Status:** v0.1.0 (alpha).

See `PUBLISHING.md` for release process.
