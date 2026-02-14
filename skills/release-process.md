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
  - Archive `name_template` must include version: `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}`
  - Binary name must include version: `{{ .ProjectName }}_v{{ .Version }}`
  - Checksum `name_template` must include version: `{{ .ProjectName }}_{{ .Version }}_SHA256SUMS`

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

- Download the released binary for your platform: `gh release download v<version> --repo Azure/mpf --pattern "azmpf_<version>_<os>_<arch>.zip"`
- Extract and verify version: `unzip <archive> && chmod +x azmpf_v<version> && ./azmpf_v<version> --version`
- **Bicep validation**: Run against a sample Bicep template (e.g., `storage-account-simple.bicep`) and verify permissions are discovered successfully
- **Terraform validation**: Initialize a sample Terraform module (e.g., `samples/terraform/aci`), then run azmpf and verify full apply/destroy cycle completes with permissions discovered
- Verify asset naming matches installation docs (format: `azmpf_<version>_<os>_<arch>.zip`)

## Step 8: Update Release Notes and Publish

- The GoReleaser configuration has `draft: true`, so the release is created as a draft
- Update the auto-generated changelog with a structured summary including sections for: New Features, Bug Fixes, Documentation, Build & Infrastructure, Refactoring, and Dependency Updates
- Publish the release: `gh release edit v<version> --draft=false --latest`
- Verify the release shows as "Latest" on GitHub

## Key Files Involved

| File | Purpose |
|------|---------|
| `.goreleaser.yml` | GoReleaser build/release configuration (targets, archives, signing, changelog) |
| `.github/workflows/release.yml` | GitHub Actions workflow triggered on tag push |
| `.github/workflows/lint.yml` | Lint workflow that validates YAML, Go, and markdown |
| `Taskfile.yml` | Task runner config — use `task lint` to run all linters locally |
| `docs/installation-and-quickstart.md` | User-facing installation docs with version-specific download URLs |
| `go.mod` | Go version and module dependencies |

## Important Considerations for Release Authors

- **Go version upgrades can break builds**: Go may drop support for certain GOOS/GOARCH pairs (e.g., Go 1.26.0 removed `windows/arm`). Always check Go release notes for dropped platform support when upgrading Go versions, and update the `ignore` list in `.goreleaser.yml` accordingly.
- **GoReleaser naming conventions must match docs**: The archive `name_template`, `binary` name, and checksum `name_template` must include `{{ .Version }}` to produce artifacts like `azmpf_0.16.0_linux_amd64.zip` that match the installation documentation.
- **Run `task lint` locally before pushing**: The CI lint workflow checks YAML, Go, Markdown, and more. Running `task lint` locally catches issues like YAML formatting before they fail in CI.
- **YAML formatting matters**: Use expanded YAML syntax (not inline `[ 'zip' ]`) to pass yamllint strict mode.
- **Validate with real Azure deployments**: Always test the released binary against actual Bicep and Terraform samples to confirm end-to-end functionality before publishing the draft release.
