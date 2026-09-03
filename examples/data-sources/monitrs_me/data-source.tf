data "monitrs_me" "self" {}

output "my_email" {
  value = data.monitrs_me.self.email
}

output "my_tier" {
  value = data.monitrs_me.self.tier
}
