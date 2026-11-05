---
page_title: "monitrs_me Data Source - monitrs"
subcategory: ""
description: |-
  Return the API key holder's identity and capabilities.
---

# monitrs_me (Data Source)

Whoami — returns the API key holder's identity, tier, scopes, and auth
mechanism. Useful for asserting the provider is authenticated correctly.

## Example Usage

```terraform
data "monitrs_me" "self" {}

output "my_email" {
  value = data.monitrs_me.self.email
}

output "my_tier" {
  value = data.monitrs_me.self.tier
}
```

## Schema

### Read-Only

- `email` (String)
- `tier` (String) One of `free`, `developer`, `starter`, `pro`, `scale`, `enterprise`.
- `auth_kind` (String) `api_key` when authenticated via a scoped key.
- `scopes` (List of String)
