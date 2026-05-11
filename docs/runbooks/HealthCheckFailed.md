# Runbook: HealthCheckFailed

| Field | Value |
|-------|-------|
| **Severity** | critical |
| **Source** | `observability/prometheus/rules/alerting.yml` |
| **Alert Group** | `alerts.uptime` |
| **Condition** | `probe_success == 0` for 1m |
| **Slack Channel** | `#alerts` |

## Summary

A blackbox probe endpoint has returned `probe_success == 0` for 1 consecutive minute, indicating the target service is unreachable or failing health checks.

## Symptoms

- Slack alert: `[FIRING] HealthCheckFailed`
- Prometheus shows `probe_success{instance="<url>"} == 0`
- Users may report HTTP 502/503/504 errors or connection timeouts

## Root Cause Analysis

1. **Check the Prometheus alert details**
   ```bash
   curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="HealthCheckFailed")'
   ```

2. **Check blackbox exporter probe results**
   ```bash
   curl -s "http://localhost:9115/probe?target=<instance>&module=http_2xx"
   ```

3. **Check if the service process is running**
   ```bash
   kubectl get pods -n <namespace> | grep <service>
   kubectl describe pod <pod-name> -n <namespace>
   ```

4. **Check recent pod logs**
   ```bash
   kubectl logs <pod-name> -n <namespace> --tail=100
   kubectl logs <pod-name> -n <namespace> --previous  # crashed pod
   ```

5. **Check service endpoints**
   ```bash
   kubectl get endpoints <service-name> -n <namespace>
   ```

6. **Check network connectivity**
   ```bash
   kubectl exec -it <debug-pod> -- curl -v http://<service>:<port>/health
   ```

## Remediation

### Pod crashed / OOMKilled
```bash
kubectl describe pod <pod-name> -n <namespace>  # look for OOMKilled
kubectl rollout restart deployment/<deployment> -n <namespace>
```

### Deployment config issue
```bash
kubectl rollout history deployment/<deployment> -n <namespace>
kubectl rollout undo deployment/<deployment> -n <namespace>  # rollback
```

### Service misconfiguration
```bash
kubectl get svc <service-name> -n <namespace> -o yaml
kubectl get ingress -n <namespace>
```

### Verify recovery
```bash
# Wait 2m then check probe
curl -s http://localhost:9090/api/v1/query?query=probe_success | jq '.data.result[] | select(.metric.instance=="<url>")'
```

## Escalation Path

1. **0–5 min**: On-call engineer investigates using steps above
2. **5–15 min**: If root cause unclear, escalate to service owner
3. **15+ min**: Page engineering lead; consider customer communication if user-facing

---
*Source: [observability/prometheus/rules/alerting.yml](../../observability/prometheus/rules/alerting.yml)*
