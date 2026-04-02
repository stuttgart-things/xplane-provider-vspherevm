# CI/CD

## Workflow Architecture

```
Push/PR to main ──> build-test ──> build-scan-image ──> release ──> publish (image + xpkg)
                                                                 └──> pages (docs)
```

## GitHub Actions Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `build-test` | Push/PR to main | Go build, unit tests, golangci-lint |
| `build-scan-image` | Push/PR to main, workflow_dispatch | Build container image via Dagger, push to ttl.sh, scan with Trivy |
| `release` | After image scan on main, workflow_dispatch | Semantic-release + publish image and xpkg to ghcr.io |
| `pages` | After release, workflow_dispatch | Deploy GitHub Pages docs |

## Build & Test

Runs on every push/PR to main with concurrency control (cancel-in-progress):

```yaml
steps:
  - checkout with submodules: true
  - setup Go from go.mod
  - go build ./...
  - go test ./... -v -count=1
  - golangci-lint v2.11.4
```

## Image Build & Scan

Uses [Dagger](https://dagger.io/) with the `stuttgart-things/blueprints/kubernetes-microservice` module:

- **Build:** Bakes image from `cluster/images/xplane-provider-vspherevm/Dockerfile`, pushes to `ttl.sh` with commit hash tag
- **Scan:** Trivy scan with severity HIGH,CRITICAL

## Release Process

Releases are automated via [semantic-release](https://semantic-release.gitbook.io/) using a reusable workflow from `stuttgart-things/github-workflow-templates`:

- `fix:` commits trigger a **patch** bump
- `feat:` commits trigger a **minor** bump
- `feat!:` or `BREAKING CHANGE:` commits trigger a **major** bump

Each release:

1. Creates a GitHub release with changelog
2. Builds and pushes container image to `ghcr.io/stuttgart-things/xplane-provider-vspherevm:<version>` + `:latest` via Dagger
3. Builds Crossplane xpkg with embedded runtime image: `crossplane xpkg build --package-root=package --embed-runtime-image=...`
4. Pushes xpkg to `ghcr.io/stuttgart-things/xplane-provider-vspherevm-xpkg:<version>` + `:latest`

## GitHub Templates

| Template | Path |
|----------|------|
| Bug Report | `.github/ISSUE_TEMPLATE/bug_report.md` |
| Feature Request | `.github/ISSUE_TEMPLATE/feature_request.md` |
| Pull Request | `.github/PULL_REQUEST_TEMPLATE.md` |

## Running Tests Locally

```bash
# Unit tests
go test ./... -v -count=1

# Build
go build ./...

# Lint (requires golangci-lint)
golangci-lint run ./...

# Docker build
docker build -f cluster/images/xplane-provider-vspherevm/Dockerfile -t xplane-provider-vspherevm:test .

# Run code generation
go run cmd/generator/main.go "$PWD"
```
