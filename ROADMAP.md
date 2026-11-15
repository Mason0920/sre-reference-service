# ROADMAP

> 프로젝트 초기에 세운 계획 기록이다. `HSLA-n`은 작업 이슈 번호이고, 상태 표기는 작성 시점 기준이다. 실제 진행은 커밋 이력을 본다.

**Goal:** Build a production-grade SRE foundation: automated CI/CD, full observability, infrastructure as code, and robust incident management.

---

## Milestone 1 — Foundation (Weeks 1–2) ✅ IN PROGRESS

Establish the project skeleton: repository, branch protection, documentation, and process definition.

| Issue | Title | Status |
|-------|-------|--------|
| HSLA-2 | Initialize GitHub repository | in_progress |
| HSLA-3 | Set up branch protection and PR requirements | in_progress |
| HSLA-4 | Create roadmap and generate initial backlog | in_progress |
| HSLA-5 | Document or establish release process | in_progress |
| HSLA-6 | Create initial project documentation | in_progress |

**Exit criteria:** Repository exists, main is protected, release process is documented, README and contribution guide are published.

---

## Milestone 2 — CI/CD Pipeline (Weeks 3–4) 🔄 IN PROGRESS

Automate lint, test, build, and deployment for every commit and PR.

| Issue | Title | Status |
|-------|-------|--------|
| HSLA-7 | Set up CI/CD pipeline | in_progress |
| HSLA-8 | Add linter and configure lint rules | in_progress |
| HSLA-9 | Add lint step to CI workflow | in_progress |
| HSLA-10 | Add test runner and initial tests | todo |
| HSLA-11 | Add test step to CI workflow | todo |
| HSLA-12 | Configure build step | todo |
| HSLA-13 | Set up deployment automation | todo |
| HSLA-14 | Add deployment smoke tests | todo |

**Exit criteria:** Every push triggers lint + test + build. Main branch auto-deploys. Smoke tests confirm live health.

---

## Milestone 3 — Observability (Weeks 5–6) 📋 BACKLOG

Instrument the application and infrastructure for full-stack visibility.

- Set up metrics collection (Prometheus + Grafana)
- Configure log aggregation (structured JSON logs → CloudWatch/ELK)
- Define SLIs and SLOs for key services
- Set up alerting rules and notification channels (Slack)
- Create operational dashboards (system health, deployment status, error rates)
- Add distributed tracing (OpenTelemetry)
- Implement uptime/health-check monitoring

**Exit criteria:** P99 latency, error rate, and saturation metrics visible in dashboards. Alerts fire on threshold breaches. On-call team has actionable context.

---

## Milestone 4 — Infrastructure as Code (Weeks 7–8) 📋 BACKLOG

Replace manual infra with reproducible, version-controlled Terraform.

- Initialize Terraform workspace (state backend: S3 + DynamoDB)
- Codify existing AWS resources (EC2, EKS, RDS, Route53, CloudFront)
- Manage environment separation (dev / staging / prod)
- Automate infra changes via CI/CD (terraform plan in PR, terraform apply on merge)
- Document infrastructure topology in `docs/INFRASTRUCTURE.md`

**Exit criteria:** All environments are Terraform-managed. Infra changes go through PR review. Drift detection runs on a schedule.

---

## Milestone 5 — Security & Compliance (Weeks 9–10) 📋 BACKLOG

Embed security scanning and secrets hygiene into the developer workflow.

- Integrate SAST tooling in CI (Semgrep, CodeQL, or Snyk)
- Add dependency vulnerability scanning (Dependabot / Trivy)
- Enforce secret detection pre-commit and in CI (gitleaks, detect-secrets)
- Migrate secrets to AWS Secrets Manager / Parameter Store
- Rotate long-lived credentials; audit IAM roles for least-privilege
- Create `docs/SECURITY.md` with disclosure policy and controls

**Exit criteria:** No plaintext secrets in repo. All PRs scanned for vulnerabilities. IAM audit completed.

---

## Milestone 6 — Incident Management (Weeks 11–12) 📋 BACKLOG

Build the operational muscle: runbooks, on-call, and post-incident learning.

- Create runbook library for top-10 alert scenarios
- Configure PagerDuty / OpsGenie on-call rotation
- Define incident severity levels and escalation paths
- Automate incident channel creation on alert trigger (Slack workflow)
- Establish postmortem template and blameless review process
- Run a game day / chaos experiment to validate readiness

**Exit criteria:** On-call rotation live. P1 runbooks reviewed. First postmortem completed.

---

## Summary Timeline

| Milestone | Weeks | Status |
|-----------|-------|--------|
| M1 Foundation | 1–2 | In Progress |
| M2 CI/CD Pipeline | 3–4 | In Progress |
| M3 Observability | 5–6 | Backlog |
| M4 Infrastructure as Code | 7–8 | Backlog |
| M5 Security & Compliance | 9–10 | Backlog |
| M6 Incident Management | 11–12 | Backlog |

_Last updated: 2026-04-30 by Product Owner_
