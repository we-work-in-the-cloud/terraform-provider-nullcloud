# NullCloud Terraform Policy example

> **Experimental, not for production.** Terraform Policy and the pinned policy-plugin framework are experimental; HashiCorp does not support this framework for production use. This example does not change the Terraform provider or register a Terraform provider function.

Two independent resource policies demonstrate:

- [Bucket region](policies/region/bucket-region.policy.hcl): configured `region` must equal `us-east`.
- [Bucket CRN](policies/crn/bucket-crn.policy.hcl): computed `crn` must pass `plugin::nullcloud::valid_crn(attrs.crn, "bucket")` in a separate policy-plugin executable.

The mock tests require no backend, credentials, Terraform state, or infrastructure. Provider acquisition does require network access on the first run: tfpolicy downloads and executes the real provider to obtain its schema, rather than using a hand-written schema.

## Prerequisites and tested versions

Install [Terraform >= 1.16.0](https://developer.hashicorp.com/terraform/install) and **tfpolicy 0.5.0** from the [official releases](https://releases.hashicorp.com/tfpolicy/0.5.0/) on your PATH (verify its published checksums). Homebrew's `tfpolicy 0.1.0` was also checked: it rejects `required_providers` with `Unsupported block type` and cannot run this schema-backed example. Check `tfpolicy version` rather than assuming the package-manager version is current. Local validation used:

| Component | Tested version |
|---|---|
| Terraform CLI | `Terraform v1.16.4` initially; `Terraform v1.17.2` in the final Homebrew run; both `on darwin_arm64` |
| tfpolicy CLI (`tfpolicy version`) | `0.5.0` |
| Go (source builds only) | `go version go1.27.4 darwin/arm64`; module requires Go >= 1.27.1 |
| Policy-plugin framework | `github.com/hashicorp/terraform-policy-plugin-framework v0.1.1` |
| Real provider schema | `registry.terraform.io/we-work-in-the-cloud/nullcloud` **0.6.0**, acquired by tfpolicy from the published provider |
| Release packaging | GoReleaser **2.20.2** |

Both policies pin provider 0.6.0: omitting the version can select the oldest published schema. A cold-cache run with an empty Terraform CLI configuration downloaded the published 0.6.0 provider and cached its real `darwin_arm64` schema. Real `tfpolicy validate` reported `Acquiring schemas for 1 provider(s) declared in required_providers...` and `Success! Terraform policies validated successfully.` Both region cases and all three CRN cases passed `tfpolicy test`, including expected violations. The registered-function Go tests and `go vet ./...` passed with framework v0.1.1, including non-string and null argument diagnostics.

Optional [HCP enforcement requires Terraform >= 1.17](https://developer.hashicorp.com/terraform/cloud-docs/policy-enforcement/manage-policy-sets/vcs/terraform-policy-vcs), not just the local CLI minimum. These requirements apply to this example only, not ordinary provider usage.

## Install a precompiled plugin (recommended)

Start in `examples/policy/`. Select a provider release **containing the policy assets** from [GitHub Releases](https://github.com/we-work-in-the-cloud/terraform-provider-nullcloud/releases). Older releases, including the schema version 0.6.0, do not contain this new plugin. The plugin shares the provider release tag/version; it is not downloaded from the Terraform Registry.

Each release produces `nullcloud-policy_<version>_<os>_<arch>.zip` separately from the provider installation ZIP. Supported tuples match the provider: `darwin/{amd64,arm64}`, `freebsd/amd64`, `linux/{386,amd64,arm,arm64,s390x}`, and `windows/{386,amd64,arm64}`. The Linux `arm` build uses GoReleaser's default GOARM=6.

Example for macOS Apple Silicon (use `OS=linux ARCH=amd64` for an HCP worker binary):

```sh
VERSION="X.Y.Z" # Replace with a published version containing nullcloud-policy assets, without v.
OS=darwin
ARCH=arm64
ASSET="nullcloud-policy_${VERSION}_${OS}_${ARCH}.zip"
SUMS="terraform-provider-nullcloud_${VERSION}_SHA256SUMS"
BASE="https://github.com/we-work-in-the-cloud/terraform-provider-nullcloud/releases/download/v${VERSION}"
curl -fLO "$BASE/$ASSET"
curl -fLO "$BASE/$SUMS"
curl -fLO "$BASE/$SUMS.sig"
```

Use the **same trusted release-signing public key as the provider** (available through the provider's Terraform Registry signing-key metadata). Import that key into GPG and independently confirm its fingerprint with the maintainer before trusting it. Do not treat a key downloaded beside an untrusted archive as proof of identity.

```sh
gpg --verify "$SUMS.sig" "$SUMS"
# Continue only if the signature is valid and belongs to the trusted provider key.
grep "  ${ASSET}$" "$SUMS" | shasum -a 256 -c -
# Continue only if the selected archive reports OK.
mkdir -p plugins/bin
# Replace the file, not an existing executable's inode (macOS caches code signatures).
rm -f plugins/bin/nullcloud-policy
unzip -j "$ASSET" nullcloud-policy -d plugins/bin
chmod +x plugins/bin/nullcloud-policy
```

The shared `terraform-provider-nullcloud_<version>_SHA256SUMS` includes both provider and plugin ZIP digests. Its detached `.sig` is generated by the existing GPG release signing step. A checksum without a trusted signature verifies file integrity, not publisher identity. These commands use POSIX shell, GPG, curl, unzip, grep, and shasum; binary consumers do not need Go.

On Windows, extract `nullcloud-policy.exe` instead and set the CRN policy's plugin source to `../../plugins/bin/nullcloud-policy.exe`. One source path selects one platform binary. Preserve executable permissions when copying or committing a Unix executable. Do not commit a local macOS binary to this provider repository; `plugins/bin/` is ignored.

## Alternative: build from source

From `examples/policy/`, using Go >= 1.27.1:

```sh
cd plugins/src/nullcloud
go mod download
go test ./...
go vet ./...
go build -o ../../bin/nullcloud-policy .
cd ../../..
```

This is an isolated module with its own `go.mod`/`go.sum`. Root `go test ./...` does not include it. The provider root dependencies and startup remain unchanged. `plugins.RegisterFunction` registers the pure Go function and `plugins.Serve()` starts the policy RPC server; Terraform provider configuration, API requests, and provider function registration are not involved.

`valid_crn(crn, expected_type)` takes two strings and returns a boolean. It accepts exactly `crn:nullcloud:<type>:<id>` where `<type>` matches `expected_type` and is one of `vpc`, `subnet`, `instance`, `loadbalancer`, `bucket`, `database`, or `cluster`. The ID must be nonempty and contain no Unicode whitespace. Prefixes and types are case-sensitive; no input is trimmed, repaired, or normalized. Malformed strings return false; non-string and null arguments produce argument diagnostics. The function validates shape only, not UUID format, existence, ownership, authorization, or region.

## Validate and test locally

From `examples/policy/`, after installing or building the matching plugin:

```sh
tfpolicy validate --policies=policies/region
tfpolicy test --policies=policies/region --tests=tests/region
tfpolicy validate --policies=policies/crn
tfpolicy test --policies=policies/crn --tests=tests/crn
```

Expect two passing region cases (`us-east` allowed, `us-west` expected to violate) and three passing CRN cases (valid bucket, mismatched VPC, malformed empty ID). `expect_failure = true` asserts a policy violation, not a parsing or plugin error. Each test file targets only its matching policy, so one rule cannot mask another rule's failure.

Run each directory explicitly. tfpolicy 0.5.0 merges files in a directory, rejects duplicate `terraform_config` blocks, and does **not** recursively discover policies: pointing it at the parent `policies/` can exit successfully without evaluating any policies. Separate directories let both policies declare their own Terraform requirement and avoid coupling the region rule to a computed CRN.

### Evaluation timing and safety

A configured region is available at plan time. A new bucket's server-generated, computed CRN is unknown during initial planning, so a rule referencing it can defer evaluation until after apply. A post-apply violation **does not undo created infrastructure**. The separate region policy does not wait for the unrelated CRN value. Local mocks establish rule outcomes with known attribute values; they do not verify hosted lifecycle timing, rollback, policy-set rollout, or enforcement permissions.

## Optional HCP delivery

For HCP Terraform >= 1.17, download the `linux/amd64` asset for remote workers, or the asset matching your agent (`linux/amd64` or `linux/arm64`). Alternatively cross-compile from `plugins/src/nullcloud/`:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../bin/nullcloud-policy .
```

Copy the policies and matching executable into **your policy VCS repository**, preserve the relative source path and executable permission, and commit that target-platform binary there (override your consumer repository's ignore rule if needed). Terraform does not fetch plugins from GitHub URLs or install them with the provider. Mixed-architecture agents cannot share one binary at one source path; use a matching policy repository/layout per execution platform. Hosted policy-set configuration and enforcement were not tested here; see [HashiCorp's plugin guidance](https://developer.hashicorp.com/terraform/policy/write/plugins).

## Release packaging checks

From the provider repository root, with GoReleaser 2.20.2, Go, and Python 3 installed:

```sh
goreleaser check
goreleaser release --snapshot --clean --skip=publish,sign
python3 examples/policy/test_release.py
```

This builds the complete provider/plugin matrix without publishing a tag or release. The packaging test verifies all 22 ZIPs across 11 platforms, executable permissions, exact archive members, separation, and every digest in the unchanged provider-named checksum manifest. Provider ZIP names and members are preserved; each plugin ZIP has the standalone `nullcloud-policy` executable (`.exe` on Windows). Snapshot signing is deliberately skipped; production releases keep the existing GPG signing path. The host-compatible snapshot executable was extracted and passed the same real `tfpolicy validate`/`test` commands above. Terraform 1.17.2 also loaded the separately packaged provider's schema successfully. CI runs nested-module tests and these packaging checks before its existing tag-only release job.
