# Architecture Overview

## System Purpose

This project delivers SRE foundational infrastructure for the `hslee` company. The primary output is a fully automated CI/CD pipeline that enforces code quality, runs tests, produces reproducible builds, and deploys to target environments with rollback capability.

## High-Level Architecture

```
Developer Workstation
        │
        │  git push / PR
        ▼
  GitHub Repository
        │
        ├─── Pull Request ──► CI Pipeline (lint → test → build)
        │                          │
        │                          │  fail → PR blocked
        │                          │  pass → PR ready for merge
        │
        └─── Merge to main ──► CD Pipeline (build → deploy → smoke test)
                                       │
                                       ├─► Staging Environment
                                       └─► Production Environment
```

## Components

### Source Control

- **Platform**: GitHub
- **Default branch**: `main` (protected)
- **Branch protection**: PR required, CI must pass, no direct pushes
- **Workflow**: See [docs/CONTRIBUTING.md](CONTRIBUTING.md) for branch naming and PR conventions

### CI Pipeline (GitHub Actions)

Triggered on: every PR and every push to `main`.

| Stage  | Tool                | Failure behaviour      |
|--------|---------------------|------------------------|
| Lint   | `golangci-lint` / `ruff` / `eslint` | Block merge |
| Test   | `go test` / `pytest` / `jest` | Block merge |
| Build  | `go build` / `docker build` | Block merge |

### CD Pipeline (GitHub Actions)

Triggered on: merge to `main`.

| Stage       | Tool / Platform     | Notes                          |
|-------------|---------------------|--------------------------------|
| Build image | Docker              | Tagged with git SHA            |
| Push image  | Container Registry  | GitHub Container Registry (GHCR) or AWS ECR |
| Deploy      | AWS EKS / EC2 / Lambda | Environment-specific config |
| Smoke test  | `curl` / custom script | Health endpoint check       |
| Rollback    | Re-deploy prior SHA | Triggered manually or on smoke failure |

### Environments

| Environment | Trigger           | URL / Target         |
|-------------|-------------------|----------------------|
| Staging     | merge to `main`   | TBD — configure in GitHub Secrets |
| Production  | manual promotion  | TBD — configure in GitHub Secrets |

> Update this table once deployment targets are confirmed.

### Infrastructure

- **Cloud**: AWS (primary)
  - EC2 / EKS for compute
  - Lambda for serverless workloads (if applicable)
  - CloudFront + Route53 for DNS and CDN
- **IaC**: Terraform (state stored in S3 + DynamoDB locking)
- **Secrets**: GitHub Secrets for CI/CD credentials; AWS Secrets Manager for application secrets

## Data Flow

1. Developer pushes code or opens PR
2. GitHub triggers the CI workflow (Actions runner)
3. Actions runner: installs dependencies → lints → tests → builds artifact
4. On PR: status checks update the PR; merge blocked on failure
5. On merge to `main`: CD workflow builds and pushes Docker image → deploys to staging → runs smoke tests
6. On smoke test failure: automated rollback to previous image SHA + alert

## Key Design Decisions

| Decision                        | Rationale                                                  |
|---------------------------------|------------------------------------------------------------|
| GitHub Actions (not Jenkins)    | No infrastructure to maintain; integrated with GitHub PRs  |
| Docker for build artifacts      | Reproducible builds across CI and production               |
| Smoke tests post-deploy         | Catch runtime failures before they become incidents        |
| Terraform for IaC               | Existing team familiarity; state management built-in       |
| Branch protection on `main`     | Prevent accidental direct pushes; enforce CI gate          |

## Monitoring and Observability

- **CI health**: Checked on weekday mornings by the DevOps routine (flaky tests, slow builds)
- **Deployment health**: Smoke test results posted as GitHub commit statuses
- **Alerts**: Configure via AWS CloudWatch or PagerDuty (TBD)

## Known Gaps / Future Work

- Production environment not yet defined — update once target platform is chosen
- Staging environment URL TBD
- Rollback automation not yet implemented — currently manual re-deploy
- Secrets rotation policy TBD
