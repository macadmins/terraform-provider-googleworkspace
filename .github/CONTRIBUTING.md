# Contributing

Thanks for helping maintain the Google Workspace provider.

## Reporting a bug

Open an issue and include:

- Terraform version (`terraform -v`)
- Provider version
- The resource or data source involved
- A minimal, sanitized configuration that reproduces it
- Full error output, with `TF_LOG=DEBUG` if the failure is in an API call (debug logs redact tokens and passwords, but review before posting)

## Changing code

1. Fork and branch from `main`.
2. `make build` and `make test` must pass. Unit tests need no credentials.
3. Acceptance tests (`make testacc`) run against a real Google Workspace tenant and need `GOOGLEWORKSPACE_CUSTOMER_ID`, `GOOGLEWORKSPACE_DOMAIN`, `GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL`, and credentials. Run the ones relevant to your change, e.g. `make testacc TESTARGS='-run=TestAccResourceGroup_'`.
4. If you change a schema or an example, run `make generate` and commit the regenerated `docs/`. Do not edit `docs/` by hand.
5. Add a line to the unreleased section of `CHANGELOG.md`.
6. Open a pull request against `main`.

`CLAUDE.md` at the repository root describes the code layout and the patterns each resource follows (CRUD shape, eventual-consistency polling, retry transport). It is a good map of the codebase for humans too.

## Modules

The Terraform modules under `modules/` have their own READMEs and YAML schemas. `terraform fmt -recursive -check ./modules/ ./examples/` runs in CI.
