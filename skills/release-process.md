# Skills: Creating a New Release of Azure/mpf

This document captures the key steps and skills used when creating a new release of the Azure MPF (Minimum Permissions Finder) repository.

## Step 1: Analyze Changes Since Last Release

- Identify the last release tag: `git tag --sort=-creatordate | head -5`
- List all commits since last release: `git log <last-tag>..HEAD --oneline`
- Filter for significant changes (features, fixes, docs): `git log <last-tag>..HEAD --oneline | grep -E "^[a-f0-9]+ (feat|fix|refactor|docs):"`
- Determine the appropriate version bump (major/minor/patch) based on the changes

## Step 2: Identify and Update Documentation

- Check `docs/installation-and-quickstart.md` for hardcoded version references in download URLs, unzip commands, and binary rename commands (Linux, macOS, Windows)
- Search all docs for old version references: `grep -rn "v<old-version>" docs/ README.md AGENTS.md`
- Update all version strings to the new release version

## Step 3: Verify Build Configuration Compatibility

- Review `.goreleaser.yml` for build targets (GOOS/GOARCH matrix)
- Check if the current Go version (from `go.mod`) has dropped support for any target platforms
  - Example: Go 1.26.0 dropped `windows/arm` support, requiring an addition to the `ignore` list in `.goreleaser.yml`
- Verify `.goreleaser.yml` naming conventions match installation docs:
  - Archive `name_template` should not include version: `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}` (version is in the release tag URL)
  - Binary name should be plain with no version suffix: `{{ .ProjectName }}` (i.e., just `azmpf`)
  - Archives use `tar.gz` (default) for Linux/macOS and `zip` for Windows via `format_overrides`
  - Checksum `name_template`: `{{ .ProjectName }}_SHA256SUMS.txt`

## Step 4: Run Linting Locally Before Committing

Run linting locally using [Task](https://taskfile.dev/) to catch issues before pushing:

```bash
# Run all linters (GitHub Actions, shell, Go, Terraform, Markdown, JavaScript, YAML)
task lint

# Or run individual linters
task yml:lint    # YAML lint (includes .goreleaser.yml)
task go:lint     # Go lint
task md:lint     # Markdown lint
task sh:lint     # Shell script lint
```

This mirrors the CI lint workflow (`.github/workflows/lint.yml`) and helps avoid failed lint runs after pushing.

## Step 5: Commit, Tag, and Push

- Commit documentation updates: `git add <files> && git commit -m "docs: update installation instructions for v<version> release"`
- Commit any build config fixes: `git add .goreleaser.yml && git commit -m "build: <description>"`
- Push commits to main: `git push origin main`
- Wait for lint workflow to pass on main before tagging
- Create an annotated tag with release notes: `git tag -a v<version> -m "Release v<version> ..."`
- Push the tag to trigger the release workflow: `git push origin v<version>`

## Step 6: Monitor Release and Lint Workflows

- The release is automated via `.github/workflows/release.yml`, triggered by tag pushes matching `v*`
- The workflow uses GoReleaser to build binaries for all supported platforms, create SBOMs, sign checksums with GPG, and create a **draft** GitHub release
- **Also monitor the lint workflow** (`.github/workflows/lint.yml`) — it runs on pushes to main and can fail independently of the release workflow
- If the release workflow fails (e.g., unsupported GOOS/GOARCH pair), or lint fails:
  1. Delete the draft release: `gh release delete v<version> --yes`
  2. Delete the remote tag: `git push origin --delete v<version>`
  3. Delete the local tag: `git tag -d v<version>`
  4. Fix the issue, commit, and push
  5. Wait for lint to pass on main
  6. Re-create and push the tag

## Step 7: Validate the Released Binary

### 7a: Install using the installation instructions

Follow the installation instructions in `docs/installation-and-quickstart.md` to download and install the released binary. This validates that the download URLs and install steps in the docs actually work with the new release.

- Verify asset naming matches installation docs (format: `azmpf_<os>_<arch>.tar.gz` for Linux/macOS, `.zip` for Windows)
- Verify the binary runs: `./azmpf --version`

### 7b: Validate with Azure deployments

Source the environment variables required for MPF:

```bash
source dev.env.export.sh
```

**Bicep validation** — run against a sample Bicep template and verify permissions are discovered:

```bash
./azmpf bicep \
  --bicepFilePath ./samples/bicep/storage-account-simple.bicep \
  --parametersFilePath ./samples/bicep/storage-account-simple-params.json \
  --jsonOutput --verbose
```

**Terraform validation** — initialize a sample Terraform module, then run azmpf and verify the full apply/destroy cycle completes:

```bash
cd ./samples/terraform/aci && $MPF_TFPATH init && cd -

./azmpf terraform \
  --workingDir $(pwd)/samples/terraform/aci \
  --varFilePath $(pwd)/samples/terraform/aci/dev.vars.tfvars \
  --jsonOutput --verbose
```

For both validations, confirm:

- The binary executes without errors
- Permissions are discovered and listed in the output
- Resources are cleaned up successfully (role definition deleted, resource group deletion initiated)

## Step 8: Update Release Notes and Publish

- The GoReleaser configuration has `draft: true`, so the release is created as a draft
- Update the auto-generated changelog with a structured summary including sections for: New Features, Bug Fixes, Documentation, Build & Infrastructure, Refactoring, and Dependency Updates
- Publish the release: `gh release edit v<version> --draft=false --latest`
- Verify the release shows as "Latest" on GitHub

## Key Files Involved

| File                                  | Purpose                                                                        |
|---------------------------------------|--------------------------------------------------------------------------------|
| `.goreleaser.yml`                     | GoReleaser build/release configuration (targets, archives, signing, changelog) |
| `.github/workflows/release.yml`       | GitHub Actions workflow triggered on tag push                                  |
| `.github/workflows/lint.yml`          | Lint workflow that validates YAML, Go, and markdown                            |
| `Taskfile.yml`                        | Task runner config — use `task lint` to run all linters locally                |
| `docs/installation-and-quickstart.md` | User-facing installation docs with version-specific download URLs              |
| `go.mod`                              | Go version and module dependencies                                             |

## Important Considerations for Release Authors

- **Go version upgrades can break builds**: Go may drop support for certain GOOS/GOARCH pairs (e.g., Go 1.26.0 removed `windows/arm`). Always check Go release notes for dropped platform support when upgrading Go versions, and update the `ignore` list in `.goreleaser.yml` accordingly.
- **GoReleaser naming conventions must match docs**: The binary name should be plain (`{{ .ProjectName }}`) with no version suffix — this is the standard convention used by most CLI tools. Archive names should also omit the version (it's already in the release tag URL). Use `tar.gz` for Linux/macOS and `zip` for Windows via `format_overrides`.
- **Run `task lint` locally before pushing**: The CI lint workflow checks YAML, Go, Markdown, and more. Running `task lint` locally catches issues like YAML formatting before they fail in CI.
- **YAML formatting matters**: Use expanded YAML syntax (not inline `[ 'zip' ]`) to pass yamllint strict mode.
- **Validate with real Azure deployments**: Always test the released binary against actual Bicep and Terraform samples to confirm end-to-end functionality before publishing the draft release.
