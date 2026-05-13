# Runbook: Deployment Rollback Procedure

**Alert name:** Rollback Triggered / Deployment Failure
**Severity:** Critical
**SLO target:** Deployment success rate ≥ 99% | Rollback rate < 2%

---

## Symptoms

- Slack alert: `[CRITICAL] Rollback triggered for production`
- GitHub Actions CD workflow failed at `smoke-test-production` job
- Newly deployed version causing errors or latency spikes
- Manual decision to revert after observing degradation post-deploy

---

## Automated Rollback (Primary Path)

The CD pipeline automatically rolls back production when smoke tests fail after deployment.

**Verify automated rollback executed:**
1. Check GitHub Actions workflow for the triggered `rollback-production` job
2. Check Slack `#deployments` channel for rollback notification
3. Verify current running image:
   ```bash
   kubectl get deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> -o jsonpath='{.spec.template.spec.containers[0].image}'
   ```
4. Confirm rollback image matches the previously deployed SHA

**If automated rollback succeeded:** proceed to [Post-Rollback Verification](#post-rollback-verification).

**If automated rollback failed or didn't trigger:** proceed to [Manual Rollback](#manual-rollback).

---

## Manual Rollback

**Time box: 5 minutes to start rollback**

### Option A — Roll back to previous Kubernetes revision (fastest):
```bash
# Check rollout history
kubectl rollout history deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>

# Roll back to previous revision
kubectl rollout undo deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>

# Watch rollout
kubectl rollout status deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> --timeout=300s
```

### Option B — Roll back to specific image SHA (precise):
```bash
# Find the last known-good image SHA from GitHub Actions logs or GHCR
PREVIOUS_IMAGE="ghcr.io/<org>/hslee:<previous-sha>"

kubectl set image deployment/hslee \
  hslee=$PREVIOUS_IMAGE \
  -n <K8S_NAMESPACE_PRODUCTION>

kubectl rollout status deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> --timeout=300s
```

### Option C — Re-trigger CD pipeline with the previous good commit:
1. Find the last good commit SHA on `main`
2. Go to GitHub → Actions → CD workflow → Run workflow → specify the commit
3. This re-runs the full pipeline with approval gate

> **Note:** Option A is fastest for immediate mitigation. Use Option B when you know the exact good image. Use Option C when you need a full pipeline validation.

---

## Post-Rollback Verification

1. **Confirm all pods are running:**
   ```bash
   kubectl get pods -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee
   ```

2. **Run health check:**
   ```bash
   curl -f https://<PRODUCTION_URL>/health
   ```

3. **Verify error rate is recovering:**
   - Check Grafana Application API Error Budget dashboard
   - Error rate should drop below 0.1% within 2–3 minutes

4. **Check p99 latency:**
   - Should return below 500ms within 2–3 minutes

5. **Post rollback notification to Slack `#incidents`:**
   ```
   Rollback complete. Production now running <previous-image-sha>.
   Error rate: <current %>, p99: <current ms>. Monitoring.
   ```

---

## Escalation Path

| Condition | Action |
|-----------|--------|
| Automated rollback triggered and succeeded | Notify `#incidents`, investigate root cause |
| Manual rollback needed, single engineer | Proceed with Option A, notify `#incidents` |
| Rollback itself fails | Escalate to engineering lead immediately |
| Both Option A and B fail | Page engineering lead + senior DevOps |
| Production down > 10 min | Declare incident, notify board |

**Escalation contacts:**
- On-call rotation: check PagerDuty schedule
- Engineering lead: Slack `#incidents` channel

---

## After Rollback Stabilizes

1. **Do not redeploy the bad commit.** Identify root cause first.
2. Create a bug fix issue and verify the fix in staging before re-promoting.
3. Review CD pipeline logs for what caused the smoke test failure.
4. If rollback rate > 2% this month: review error budget and deployment process.
5. Open postmortem issue if:
   - Production was degraded for > 5 minutes
   - Manual intervention was required
   - Automated rollback failed

---

## Rollback Decision Criteria

| Signal | Decision |
|--------|----------|
| Smoke test failure after deploy | Automated rollback triggers — verify it ran |
| Error rate > 1% after deploy | Manual rollback immediately |
| p99 > 1s after deploy | Manual rollback immediately |
| Error rate 0.1–1%, stable | Investigate for 5 min, then roll back if not improving |
| Any user-visible impact > 5 min | Roll back, then investigate |
