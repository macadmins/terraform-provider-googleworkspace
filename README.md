# Terraform Provider Google Workspace

[![Registry](https://img.shields.io/badge/registry-macadmins%2Fgoogleworkspace-623CE4)](https://registry.terraform.io/providers/macadmins/googleworkspace)
[![Release](https://img.shields.io/github/v/release/macadmins/terraform-provider-googleworkspace)](https://github.com/macadmins/terraform-provider-googleworkspace/releases)
[![License](https://img.shields.io/github/license/macadmins/terraform-provider-googleworkspace)](LICENSE)

Community-maintained Terraform provider for Google Workspace: users, groups (static and dynamic), org units, domains, roles, custom schemas, Gmail send-as aliases, and Chrome policy.

Maintained under the [Mac Admins Open Source](https://github.com/macadmins) organization. It continues the [HashiCorp provider](https://github.com/hashicorp/terraform-provider-googleworkspace) (archived upstream) by way of the [`vdesouza`](https://github.com/vdesouza/terraform-provider-googleworkspace) fork, whose full history and release line (1.0.0 to 1.4.0) live in this repository.

## Using the provider

```hcl
terraform {
  required_providers {
    googleworkspace = {
      source  = "macadmins/googleworkspace"
      version = ">= 1.5.0"
    }
  }
}

provider "googleworkspace" {
  credentials             = "/path/to/service-account-key.json"
  customer_id             = "A01b123xz"
  impersonated_user_email = "admin@example.com"
}
```

Full provider documentation, including authentication and OAuth scope setup, is on the [Terraform Registry](https://registry.terraform.io/providers/macadmins/googleworkspace/latest/docs).

### Migrating from `vdesouza/googleworkspace` or `hashicorp/googleworkspace`

The provider binary is unchanged; only its registry address moved. Update `source` in `required_providers`, then point existing state at the new address:

```sh
terraform state replace-provider registry.terraform.io/vdesouza/googleworkspace registry.terraform.io/macadmins/googleworkspace
# or, from the archived upstream:
terraform state replace-provider registry.terraform.io/hashicorp/googleworkspace registry.terraform.io/macadmins/googleworkspace

terraform init -upgrade
```

Note `hashicorp/googleworkspace` stopped at 0.7.0; review [CHANGELOG.md](CHANGELOG.md) for the changes between 0.7.0 and the current release before upgrading from it.

## Project Status

Actively maintained. The most exercised areas are Dynamic Groups, Chrome Policy resources, and the companion modules below. Other resources were inherited from upstream and have had less recent attention; test changes in a non-production Google Workspace tenant first.

See [CHANGELOG.md](CHANGELOG.md) for release-by-release notes.

## Requirements

- Terraform >= 1.4
- Go >= 1.24 (for development)
- Access to a Google Workspace environment

## Build

```sh
make build
```

## Test

```sh
make test      # unit tests
make testacc   # acceptance tests; require Google Workspace credentials and env vars
```

## Generate Documentation

```sh
make generate
```

Files under `docs/` are generated. Edit resource schemas and `examples/`, then run `make generate`.

## Companion Modules

This repository also ships YAML-driven Terraform modules that compose provider resources into higher-level Chrome management workflows. They live under [`modules/`](modules/) and are consumed via a Git source pinned to a release tag:

```hcl
module "chrome_policies" {
  source = "git::https://github.com/macadmins/terraform-provider-googleworkspace.git//modules/policies?ref=v1.5.0"
  # ...
}
```

| Module | Purpose |
| --- | --- |
| [`variables`](modules/variables/) | Centralized YAML variable substitution for the other modules. |
| [`groups`](modules/groups/) | Static and dynamic Google Workspace groups from YAML. |
| [`assets`](modules/assets/) | File uploads to Chrome Policy storage (wallpapers, avatars, ToS). |
| [`policies`](modules/policies/) | Chrome policies for groups and OUs, with asset reference resolution. |
| [`extensions`](modules/extensions/) | A variation on `policies` for Chrome extensions, Android apps, and web apps deployed to groups or OUs. |
| [`group_priority`](modules/group_priority/) | Resolves ordering when multiple groups assign overlapping policies/extensions. |

See [`modules/README.md`](modules/README.md) for the dependency graph and reference configurations.

## Releasing

Releases are cut by pushing a `v*` tag. GitHub Actions runs GoReleaser, signs the checksums with the organization's GPG key, and publishes a GitHub Release, which the Terraform Registry ingests automatically.

## Contributing

Contributions and bug reports are welcome; see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

[Mozilla Public License 2.0](LICENSE).
