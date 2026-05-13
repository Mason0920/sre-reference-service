# Runbook: High Error Rate (5xx Spike)

**Alert name:** Fast Burn (1h) / Fast Burn (6h)
**Severity:** Critical
**SLO target:** Error rate < 0.1% | Availability ≥ 99.9%
**Error budget:** 43.2 min/month

---

## Symptoms

- PagerDuty/Slack alert: `[CRITICAL] Error budget burn rate > 14.4x` or `> 6x`
- HTTP 5xx responses increasing in Grafana → Application API Error Budget dashboard
- Users reporting errors or service unavailability
- Prometheus query showing elevated error rate:
  ```
  sum(rate(http_requests_total{job="hslee",status=~"5.."}[5m])) / sum(rate(http_requests_total{job="hslee"}[5m]))
  ```

---

## Diagnosis Steps

**Time box: 3 minutes**

1. **Check current error rate:**
   ```bash
   kubectl top pods -n <K8S_NAMESPACE_PRODUCTION>
   kubectl get pods -n <K8S_NAMESPACE_PRODUCTION>
   ```

2. **Check pod logs for errors:**
   ```bash
   kubectl logs -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee --tail=100 | grep -E "ERROR|FATAL|panic"
   ```

3. **Check recent deployments** — did this start after a deploy?
   ```bash
   kubectl rollout history deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>
   ```

4. **Check pod health:**
   ```bash
   kubectl describe pods -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee | grep -A5 "Events:"
   ```

5. **Check resource pressure:**
   ```bash
   kubectl describe nodes | grep -A5 "Conditions:"
   ```

6. **Grafana:** Open Application API Error Budget dashboard and check burn rate timeline to identify when errors started.

---

## Mitigation Steps

### If caused by a recent deployment:
→ **Follow [deployment-rollback.md](./deployment-rollback.md)**

### If caused by upstream dependency (DB, external API):
1. Identify the failing dependency from logs
2. Check if dependency has its own status page or alerts
3. If dependency is down: implement circuit breaking or return graceful degraded response
4. Notify dependency team and track their ETA

### If caused by resource exhaustion (OOM, CPU throttling):
1. Scale up:
   ```bash
   kubectl scale deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> --replicas=<current+2>
   ```
2. Monitor pod restarts:
   ```bash
   kubectl get pods -n <K8S_NAMESPACE_PRODUCTION> -w
   ```

### If caused by bad config or environment variable:
1. Check current configmap/secrets:
   ```bash
   kubectl get configmap -n <K8S_NAMESPACE_PRODUCTION>
   kubectl describe configmap/hslee-config -n <K8S_NAMESPACE_PRODUCTION>
   ```
2. Roll back config change or correct value

### General mitigation — stop the bleeding:
- If error rate > 10% and not improving within 5 minutes: **immediately roll back**
- If error rate 1–10%: continue diagnosis for up to 10 minutes before rolling back

---

## Escalation Path

| Condition | Action |
|-----------|--------|
| Error rate < 1%, improving | Monitor, no escalation |
| Error rate 1–10%, not improving after 5min | Escalate to senior DevOps on-call |
| Error rate > 10% | Immediately roll back + page engineering lead |
| Error budget < 10% remaining | Declare incident, notify board |
| Root cause unknown after 15min | Page engineering lead, open incident bridge |

**Escalation contacts:**
- On-call rotation: check PagerDuty schedule
- Engineering lead: Slack `#incidents` channel
- Status page: update if user-visible impact > 5 min

---

## Post-Incident

- Open a postmortem issue within 24h if:
  - Error budget consumed > 10%
  - Incident duration > 30 min
  - User-visible impact occurred
- Review error budget status in Grafana
- Update alert thresholds if they fired too late or too early
