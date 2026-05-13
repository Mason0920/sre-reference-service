# Rollback Procedures

## Overview

The CD pipeline automatically rolls back when a post-deploy smoke test fails. This document covers both automated and manual rollback procedures for staging and production.

## Automated Rollback

### How It Works

1. Deploy job records the **previous image** before every deploy (`kubectl get deployment … -o jsonpath=…`)
2. If the smoke test job fails, the `rollback` / `rollback-production` job runs automatically
3. The rollback job runs `kubectl set image` with the recorded previous image, then waits for rollout to stabilize
4. Slack notification is sent with the failed SHA and the image rolled back to

### Trigger Conditions

| Environment | Trigger | Rollback target |
|---|---|---|
| Staging | `smoke-test` fails | `deploy-staging.outputs.previous_image` |
| Production | `smoke-test-production` fails | `deploy-production.outputs.previous_image` |

The rollback job **does not run** if `previous_image` is empty (i.e., first-ever deploy to that namespace).

---

## Manual Rollback

Use manual rollback when:
- The automated rollback did not trigger (e.g., first deploy scenario)
- You need to roll back to a specific earlier SHA (not just previous)
- The CI/CD pipeline itself is unavailable

### Prerequisites

```bash
# AWS credentials configured
aws configure

# kubeconfig for target cluster
aws eks update-kubeconfig \
  --name <EKS_CLUSTER_NAME> \
  --region <AWS_REGION>
```

### Roll Back to Previous Image

```bash
# 1. Find the currently running image
kubectl get deployment hslee \
  --namespace <NAMESPACE> \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# 2. List recent ReplicaSets to find a known-good image
kubectl get replicaset \
  --namespace <NAMESPACE> \
  --sort-by='.metadata.creationTimestamp' \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[0].image}{"\n"}{end}'

# 3. Roll back using kubectl rollout undo (reverts to previous ReplicaSet)
kubectl rollout undo deployment/hslee \
  --namespace <NAMESPACE>

# 4. Verify rollback
# Timeout: 120s for staging, 300s for production
kubectl rollout status deployment/hslee \
  --namespace <NAMESPACE> \
  --timeout=120s   # replace with 300s for production
```

### Roll Back to a Specific SHA

```bash
# Replace <SHA> with the full commit SHA of the known-good image
TARGET_IMAGE="ghcr.io/<org>/<repo>:<SHA>"

kubectl set image deployment/hslee \
  hslee="${TARGET_IMAGE}" \
  --namespace <NAMESPACE>

kubectl rollout status deployment/hslee \
  --namespace <NAMESPACE> \
  --timeout=120s   # replace with 300s for production
```

### Verify Health After Manual Rollback

```bash
# Check pod status
kubectl get pods --namespace <NAMESPACE> -l app=hslee

# Check health endpoint
curl -f <ENVIRONMENT_URL>/health

# Check version endpoint
curl -f <ENVIRONMENT_URL>/version
```

---

## Environment Variables Reference

| Variable | Staging value (GitHub var) | Production value (GitHub var) |
|---|---|---|
| Cluster name | `EKS_CLUSTER_NAME_STAGING` | `EKS_CLUSTER_NAME_PRODUCTION` |
| K8s namespace | `K8S_NAMESPACE_STAGING` | `K8S_NAMESPACE_PRODUCTION` |
| App URL | `STAGING_URL` | `PRODUCTION_URL` |
| AWS region | `AWS_REGION` | `AWS_REGION` |

---

## Post-Rollback Checklist

- [ ] Smoke test passes against the rolled-back environment
- [ ] Notify the team in Slack of the rollback and affected SHA
- [ ] Open an incident issue to track root-cause analysis
- [ ] Do **not** merge or re-deploy the failing SHA until the root cause is fixed
- [ ] Update the failed deploy's GitHub Actions run with a summary comment

---

## GitHub Environment Protection Rules

Production deploys require manual approval from an authorized reviewer before the `deploy-production` job starts. This is enforced via the **`production` GitHub environment**.

To configure:
1. Go to **Settings → Environments → production**
2. Add **Required reviewers** (at least one person must approve before deploy proceeds)
3. Optionally enable **Wait timer** (e.g., 5 minutes) to allow last-minute cancellation

Staging deploys are automatic (no approval gate) to enable fast iteration.
