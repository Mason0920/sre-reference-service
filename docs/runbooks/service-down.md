# Runbook: Service Down / Health Check Failure

**Alert name:** Health Check Failure / Availability SLO Breach
**Severity:** Critical
**SLO target:** Health check availability ≥ 99.95% | API availability ≥ 99.9%
**Error budget:** Health check 21.6 min/month | API 43.2 min/month

---

## Symptoms

- External uptime monitor alert: `GET /health returning non-200`
- Kubernetes liveness probe failing → pod restart loop
- Grafana: availability dropping below 99.9%
- Prometheus: `sum(rate(http_requests_total{job="hslee",handler="/health",status="200"}[5m])) / sum(rate(http_requests_total{job="hslee",handler="/health"}[5m]))` near 0
- Slack alert: `[CRITICAL] Service hslee health check failing`
- Users unable to reach the service at all

---

## Diagnosis Steps

**Time box: 2 minutes** (service is down — move fast)

1. **Check pod status immediately:**
   ```bash
   kubectl get pods -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee
   ```
   - Look for: `CrashLoopBackOff`, `OOMKilled`, `Error`, `Pending`, low `READY` count

2. **Check pod logs for crash reason:**
   ```bash
   kubectl logs -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee --tail=50 --previous
   # or for currently running (crashing) pod:
   kubectl logs -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee --tail=50
   ```

3. **Check pod events:**
   ```bash
   kubectl describe pods -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee | grep -A10 "Events:"
   ```

4. **Check if it's a deployment issue:**
   ```bash
   kubectl rollout history deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>
   kubectl rollout status deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>
   ```

5. **Check node health:**
   ```bash
   kubectl get nodes
   kubectl describe nodes | grep -E "Ready|MemoryPressure|DiskPressure|PIDPressure"
   ```

6. **Verify service and ingress routing:**
   ```bash
   kubectl get svc -n <K8S_NAMESPACE_PRODUCTION>
   kubectl get ingress -n <K8S_NAMESPACE_PRODUCTION>
   ```

---

## Mitigation Steps

### Scenario A — CrashLoopBackOff (application crashing):

1. Check logs for panic/fatal error (step 2 above)
2. If caused by bad config or env var: fix or roll back config
3. If caused by recent deployment:
   → **Follow [deployment-rollback.md](./deployment-rollback.md)**
4. If unknown cause: roll back immediately, investigate after

### Scenario B — All pods terminated / scaled to 0:

1. Check if someone accidentally scaled down:
   ```bash
   kubectl get deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> -o jsonpath='{.spec.replicas}'
   ```
2. If 0 replicas: scale up:
   ```bash
   kubectl scale deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> --replicas=2
   ```

### Scenario C — OOMKilled (out of memory):

1. Confirm in pod events: `OOMKilled`
2. Temporarily increase memory limit:
   ```bash
   kubectl patch deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> \
     --patch '{"spec":{"template":{"spec":{"containers":[{"name":"hslee","resources":{"limits":{"memory":"512Mi"}}}]}}}}'
   ```
3. Monitor pod stability
4. File issue to address memory leak root cause

### Scenario D — Node failure / scheduling issue:

1. Check if pods are stuck in `Pending`:
   ```bash
   kubectl describe pods -n <K8S_NAMESPACE_PRODUCTION> | grep -A5 "Events:"
   ```
2. If node is NotReady:
   ```bash
   kubectl cordon <failing-node>
   kubectl drain <failing-node> --ignore-daemonsets --delete-emptydir-data
   ```
3. Verify pods reschedule to healthy nodes

### Scenario E — Networking / ingress issue (pods healthy but unreachable):

1. Verify pods respond locally:
   ```bash
   kubectl exec -it <pod-name> -n <K8S_NAMESPACE_PRODUCTION> -- wget -qO- localhost:8080/health
   ```
2. If pods respond locally but not externally: ingress/load balancer issue
3. Check ingress controller logs:
   ```bash
   kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx --tail=50
   ```

---

## Emergency Restart (Last Resort)

If all else fails and you need to restore service immediately:

```bash
# Force restart all pods (rolling restart)
kubectl rollout restart deployment/hslee -n <K8S_NAMESPACE_PRODUCTION>

# Watch until ready
kubectl rollout status deployment/hslee -n <K8S_NAMESPACE_PRODUCTION> --timeout=300s
```

---

## Verification After Recovery

1. **Health check returns 200:**
   ```bash
   curl -f https://<PRODUCTION_URL>/health
   ```
2. **All pods in Running/Ready state:**
   ```bash
   kubectl get pods -n <K8S_NAMESPACE_PRODUCTION> -l app=hslee
   ```
3. **Error rate returning to < 0.1%** in Grafana
4. **p99 latency returning to < 500ms** in Grafana

---

## Escalation Path

| Condition | Action |
|-----------|--------|
| Service down < 2 min, cause identified | Mitigate, notify `#incidents` |
| Service down > 2 min, cause unknown | Escalate to senior DevOps on-call |
| Service down > 5 min | Page engineering lead, update status page |
| Service down > 10 min | Declare incident, notify board |
| Node-level failure | Escalate to infrastructure team |

**Escalation contacts:**
- On-call rotation: check PagerDuty schedule
- Engineering lead: Slack `#incidents` channel
- Status page: update immediately for outages > 5 min

**Error budget impact note:** At 99.95% SLO, the health check error budget is only 21.6 min/month. Any outage > 5 minutes consumes > 20% of monthly budget. Escalate aggressively.

---

## Post-Incident

- Open postmortem issue for any outage > 5 minutes
- Review liveness/readiness probe configuration if pod restarts caused cascading failure
- Review error budget consumption and adjust freeze policy if needed (see `docs/SLO.md`)
