# CI/CD Pipeline

## Pipeline Overview

GitHub Actions powers both CI and CD. CI runs on every PR and push to `main`; CD triggers automatically when CI succeeds on `main` (via `workflow_run`). Docker images are pushed to GitHub Container Registry (GHCR) and deployed to AWS EKS.

```
PR / push ──► CI (.github/workflows/ci.yml)
                   lint → test → build
                         │ all pass on main
                         ▼
              CD (.github/workflows/cd.yml)
                   build-push → deploy-staging → smoke-test
                                                      │ fail
                                                      ▼
                                                  rollback
```

## CI Workflow

File: `.github/workflows/ci.yml`  
Runs on: PRs and pushes to `main`

| Stage | Tool | Failure behaviour |
|-------|------|-------------------|
| Lint  | `golangci-lint` v1.64.8 | Blocks merge |
| Test  | `go test -race -count=1` | Blocks merge |
| Build | `go build ./...` | Blocks merge |

## CD Workflow

File: `.github/workflows/cd.yml`  
Runs on: `workflow_run` completion of CI on `main` (success only)

| Stage | What it does |
|-------|--------------|
| Guard | Validates CI passed; exports `sha` and image tag |
| Build & Push | Builds multi-stage Docker image; pushes to GHCR tagged with git SHA and `latest` |
| Deploy → Staging | Updates `hslee` Deployment on the staging EKS cluster via `kubectl set image`; waits for rollout |
| Smoke Test | `curl` health-check loop (5 attempts × 10 s) against `$STAGING_URL/health` |
| Rollback | If smoke test fails: re-sets the Deployment to the previous image SHA |

## Environment Variables & Secrets

Configure in GitHub → Settings → Secrets and variables → Actions.

| Name | Kind | Purpose |
|------|------|---------|
| `AWS_ACCESS_KEY_ID` | Secret | IAM key for EKS/ECR access |
| `AWS_SECRET_ACCESS_KEY` | Secret | IAM secret |
| `AWS_REGION` | Variable | e.g. `ap-northeast-2` |
| `EKS_CLUSTER_NAME_STAGING` | Variable | Staging EKS cluster name |
| `K8S_NAMESPACE_STAGING` | Variable | Kubernetes namespace for staging |
| `STAGING_URL` | Variable | Base URL for smoke-test health check |

The CD workflow uses `GITHUB_TOKEN` (auto-injected) to push images to GHCR — no extra secret needed for the registry.

## Deployment Targets

| Environment | Trigger | Registry tag |
|-------------|---------|--------------|
| Staging | Automatic on every CI-passing main merge | `ghcr.io/<org>/<repo>:<git-sha>` |
| Production | Manual approval gate (after staging smoke test passes) | `ghcr.io/<org>/<repo>:<git-sha>` |

## Rollback Procedure

### Automatic (smoke test failure)

The `rollback` job runs automatically when `smoke-test` fails:

1. Reads `previous_image` output recorded before the deploy.
2. Runs `kubectl set image` to revert the Deployment.
3. Waits for rollout to complete.
4. Prints a summary (failed SHA + rolled-back image) to the GitHub Actions job summary.

### Manual rollback

```bash
# Find the last-known-good SHA from GHCR or git log
GOOD_SHA=<commit-sha>
IMAGE="ghcr.io/<org>/<repo>:${GOOD_SHA}"

kubectl set image deployment/hslee hslee="${IMAGE}" \
  --namespace <staging-namespace>

kubectl rollout status deployment/hslee \
  --namespace <staging-namespace> --timeout=120s
```

Alternatively, use `kubectl rollout undo`:

```bash
kubectl rollout undo deployment/hslee --namespace <staging-namespace>
```

## Docker Image

Dockerfile uses a two-stage build:

1. **Builder** (`golang:1.21-alpine`): compiles the binary with `CGO_ENABLED=0`.
2. **Final** (`distroless/static-debian12:nonroot`): minimal, non-root runtime image.

Image labels include `org.opencontainers.image.revision` (git SHA) and `.source` (repo URL) for traceability.
