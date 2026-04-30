# API Reference

> **Note:** This project is primarily infrastructure/SRE tooling. The "API" documented here covers internal tooling interfaces: the Makefile task runner, CI/CD workflow inputs, and any health/status endpoints exposed by deployed services.
>
> Update this document when application-specific APIs are added.

## Makefile Interface

The `Makefile` is the primary developer interface. All commands should be run from the project root.

### Core Tasks

| Target         | Description                                    | Example                    |
|----------------|------------------------------------------------|----------------------------|
| `make install` | Install project dependencies                   | `make install`             |
| `make lint`    | Run linter; non-zero exit on errors            | `make lint`                |
| `make test`    | Run full test suite                            | `make test`                |
| `make build`   | Build production artifacts                     | `make build`               |
| `make fmt`     | Auto-format source code                        | `make fmt`                 |
| `make clean`   | Remove build artifacts and temporary files     | `make clean`               |

### Extended Tasks (add as pipeline matures)

| Target             | Description                                        |
|--------------------|----------------------------------------------------|
| `make docker-build`| Build Docker image locally                         |
| `make docker-run`  | Run Docker image locally                           |
| `make smoke-test`  | Run smoke tests against a running environment      |
| `make deploy`      | Deploy to staging (requires env vars)              |

## CI/CD Workflow Inputs

### CI Workflow (`.github/workflows/ci.yml`)

Triggered automatically. No manual inputs required.

**Required secrets** (set in GitHub repository settings):

| Secret             | Description                                    |
|--------------------|------------------------------------------------|
| _(none yet)_       | Add secrets here as they are configured        |

**Outputs / status checks:**

| Check    | Status context              | Passes when…                  |
|----------|-----------------------------|-------------------------------|
| Lint     | `ci / lint`                 | Zero lint errors              |
| Test     | `ci / test`                 | Zero test failures            |
| Build    | `ci / build`                | Build exits 0                 |

### CD Workflow (`.github/workflows/cd.yml`)

Triggered on merge to `main`.

**Required secrets:**

| Secret                  | Description                                     |
|-------------------------|-------------------------------------------------|
| `AWS_ACCESS_KEY_ID`     | AWS credentials for deployment (or use OIDC)    |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key                                  |
| `DEPLOY_TARGET`         | Deployment target URL or ARN                    |

**Workflow inputs (manual trigger):**

| Input          | Type   | Default    | Description                             |
|----------------|--------|------------|-----------------------------------------|
| `environment`  | string | `staging`  | `staging` or `production`               |
| `image_tag`    | string | _(latest)_ | Specific image tag to deploy (for rollback) |

## Health Endpoints

> Add this section when the application exposes HTTP services.

### Standard Health Check

All deployed services should expose:

```
GET /healthz
```

**Response (healthy):**

```json
{
  "status": "ok",
  "version": "1.2.3",
  "commit": "abc1234"
}
```

**Response codes:**

| Code | Meaning                          |
|------|----------------------------------|
| 200  | Service is healthy               |
| 503  | Service is unhealthy / degraded  |

### Readiness Check

```
GET /readyz
```

Returns `200` when the service is ready to serve traffic (dependencies connected, warm-up complete). Used by Kubernetes readiness probes.

## Internal Tooling APIs

### Paperclip Workflow

This project is managed via [Paperclip](https://paperclip.ing). Key interactions:

| Action                  | Paperclip endpoint                                |
|-------------------------|---------------------------------------------------|
| View issues             | `/HSLA/issues`                                    |
| View this project       | `/HSLA/projects/hslee`                            |
| CI/CD goal              | `/HSLA/goals` → "Set up CI/CD pipeline"           |

### GitHub CLI Reference

Common `gh` commands used in this project's workflow:

```bash
# Open a PR
gh pr create --title "feat: description" --body "$(cat body.md)"

# Check CI status
gh pr checks

# Merge a PR
gh pr merge <number> --merge

# View recent CI runs
gh run list --limit 10

# Watch a CI run
gh run watch <run-id>
```

## Changelog

> Document breaking API or interface changes here as the project evolves.

| Version | Date       | Change                              |
|---------|------------|-------------------------------------|
| 0.1.0   | 2026-04-30 | Initial documentation scaffold      |
