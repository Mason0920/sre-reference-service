# Runbook: High Latency / p99 SLO Breach

**Alert name:** p99 Latency Breach
**Severity:** Warning
**SLO target:** p99 latency < 500ms
**Health check SLO:** p99 < 100ms

---

## Symptoms

- Slack/PagerDuty alert: `[WARNING] p99 latency > 500ms for > 5 minutes`
- Grafana Application API dashboard shows p99 climbing above threshold
- Users reporting slow responses
- Prometheus query elevated:
  ```
  histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="hslee"}[5m])) > 0.5
  ```

---

## Diagnosis Steps

**Time box: 3 minutes**

1. **Verify the alert with current metrics:**
   ```bash
   # Check pod resource usage
   kubectl top pods -n <K8S_NAMESPACE_PRODUCTION>
   ```

2. **Check if specific endpoints are slow** (look at per-handler metrics in Grafana):
   - Is it all endpoints or just one?
   - Is `/health` also slow? (SLO: p99 < 100ms) — if yes, this is systemic

3. **Check CPU/memory pressure on nodes:**
   ```bash
   kubectl top nodes
   kubectl describe nodes | grep -A10 "Allocated resources:"
   ```

4. **Check for slow database queries or external calls in logs:**
   ```bash
   kubectl logs -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee --tail=200 | grep -E "slow|timeout|duration|latency"
   ```

5. **Check if this correlates with traffic spike:**
   - Compare request rate in Grafana vs. baseline
   - Check if p50 latency is also elevated (systemic) vs. only p99 (outliers)

6. **Check for GC pauses or runtime issues** (Go application):
   ```bash
   kubectl logs -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee --tail=200 | grep -i "gc\|goroutine\|heap"
   ```

7. **Check recent deployment:**
   ```bash
   kubectl rollout history deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>
   ```

---

## Mitigation Steps

### If caused by traffic spike (scaling issue):
1. Scale horizontally:
   ```bash
   kubectl scale deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> --replicas=<current+2>
   ```
2. Watch p99 improve over the next 2 minutes
3. Set HPA if not already configured

### If caused by a slow dependency (DB, cache, external API):
1. Identify from logs which call is slow
2. Check dependency health independently
3. If dependency is degraded:
   - Enable caching / serve stale if possible
   - Circuit break if available
   - Contact dependency owner

### If caused by a recent deployment:
→ **Follow [deployment-rollback.md](./deployment-rollback.md)**

### If caused by resource starvation (CPU throttling):
1. Check limits:
   ```bash
   kubectl describe deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> | grep -A5 "Limits:"
   ```
2. Temporarily increase CPU limit via patch or roll back to last known-good manifest

### If p99 > 1s (severe breach):
- Treat as critical — escalate immediately
- Consider rollback without waiting for root cause

---

## Decision Thresholds

| p99 Latency | Duration | Action |
|-------------|----------|--------|
| 500ms – 1s | < 5 min | Monitor, investigate |
| 500ms – 1s | > 5 min | Scale up, investigate root cause |
| 1s – 2s | Any | Scale up immediately + escalate |
| > 2s | Any | Roll back immediately |

---

## Escalation Path

| Condition | Action |
|-----------|--------|
| p99 500ms–1s, improving after scale | No escalation, continue monitoring |
| p99 > 1s, not improving after 5min | Escalate to senior DevOps on-call |
| p99 > 2s | Immediate rollback + page engineering lead |
| Root cause unknown after 15min | Page engineering lead, open incident bridge |

**Escalation contacts:**
- On-call rotation: check PagerDuty schedule
- Engineering lead: Slack `#incidents` channel

---

## Post-Incident

- Document root cause in Slack `#incidents`
- If p99 exceeded 2s for > 10 minutes: open postmortem issue
- Review if HPA configuration needs tuning
- Review if latency SLO threshold needs adjustment
