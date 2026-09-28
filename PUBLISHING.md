# Publishing `terraform-provider-monitrs` to the Terraform Registry

Follows HashiCorp's [publishing guide][pub].

[pub]: https://developer.hashicorp.com/terraform/registry/providers/publishing

## One-time setup (operator)

### 1. Extract to a dedicated repo

The Terraform Registry REQUIRES the provider to live in a repo named
`terraform-provider-<name>`. This provider currently lives in
`monit.rs/terraform/` for fast iteration. When ready to publish:

```bash
# From the monit.rs repo root
git subtree split --prefix=terraform -b split-terraform
# Create the new repo on GitHub: Monitrs/terraform-provider-monitrs
git remote add tf-repo git@github.com:Monitrs/terraform-provider-monitrs.git
git push tf-repo split-terraform:main
```

The `go.mod` module path is already `github.com/Monitrs/terraform-provider-monitrs`
so no changes required after extraction.

### 2. Generate a GPG key for signing

```bash
gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Name-Real: monit.rs Terraform Provider
Name-Email: support@monit.rs
Expire-Date: 0
%commit
EOF

# Get the fingerprint:
gpg --list-secret-keys --keyid-format LONG
# Export public key for the Registry:
gpg --armor --export <FINGERPRINT> > monitrs-tf-provider-public.asc
```

### 3. Register on the Registry

There is **no terraform.io or HCP account to create** — the Registry
authenticates with GitHub OAuth only. Per HashiCorp's guide: *"Before
publishing a provider, you must first sign in to the Terraform Registry with a
GitHub account."*

1. Go to https://registry.terraform.io and **Sign in with GitHub**, authorising
   the Terraform Registry OAuth app.
   Sign in as your GitHub **user** (e.g. `Zenofex`), not as the org. The
   account needs **admin** on the provider repo — org ownership alone is not
   what the Registry checks. Verify under GitHub Settings → Applications →
   Authorized OAuth Apps → Terraform Registry.
2. Register the signing key under **User Settings → Signing Keys**, pasting the
   ASCII-armored public key from step 2 above. Depending on the flow the
   publish wizard may also offer to take it inline; it is the same key either
   way, and it must match `GPG_FINGERPRINT` in the release workflow or imports
   fail signature verification.
3. Go to https://registry.terraform.io/publish/provider, select
   `terraform-provider-monitrs`, and confirm.

Registry prerequisites, all of which this repo already satisfies: a public repo
named `terraform-provider-{name}` in lowercase; `docs/` with an overview plus a
page per resource and data source; a GitHub release on a `vX.Y.Z` tag; and a
registered GPG public key.

### 4. Add secrets to the GitHub repo

In `Monitrs/terraform-provider-monitrs` (the extracted repo) → Settings → Secrets:

- `GPG_PRIVATE_KEY`: `gpg --export-secret-keys --armor <FINGERPRINT>`
- `PASSPHRASE`: (empty if you used `%no-protection` above, else the passphrase)

## Per-release flow

```bash
# In the extracted terraform-provider-monitrs repo:
git tag v0.1.0
git push --tags

# Create a GitHub Release from the tag (mark as pre-release for v0.x)
# The workflow is .github/workflows/release.yml in this directory, so the
# subtree split in step 1 carries it into the extracted repo. It fires on
# tags matching v*, imports the signing key, and passes its fingerprint to
# GoReleaser as GPG_FINGERPRINT.
# GoReleaser workflow auto-builds signed zips for 11 os/arch combos + publishes
# to the release. The Registry polls the repo every ~10 minutes and imports
# new releases automatically.

# Verify at:
# https://registry.terraform.io/providers/Monitrs/monitrs/latest
```

## Common failure modes

- **"provider not found" in `terraform init`**: give the Registry 10-30 min to import a new release
- **Signature verification fails**: GPG_FINGERPRINT env in the workflow must match the one Registry has on file
- **Zip name mismatch**: `.goreleaser.yml` binary name MUST be `{{ .ProjectName }}_v{{ .Version }}` — the `v` prefix is required by Registry
