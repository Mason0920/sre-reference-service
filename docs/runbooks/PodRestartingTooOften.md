# Runbook: PodRestartingTooOften

| Field | Value |
|-------|-------|
| **Severity** | warning |
| **Source** | `observability/prometheus/rules/alerting.yml` |
| **Alert Group** | `alerts.kubernetes` |
| **Condition** | `increase(kube_pod_container_status_restarts_total[5m]) > 3` for 1m |
| **Slack Channel** | `#alerts` |

## Summary

A Kubernetes pod restarted more than 3 times in the last 5 minutes. Repeated restarts indicate a crash loop, OOM kill, or application startup failure.

## Symptoms

- Slack alert: `[FIRING] PodRestartingTooOften`
- `kubectl get pods` shows `CrashLoopBackOff` or high restart count
- Application unavailable intermittently

## Root Cause Analysis

1. **Identify the restarting pod**
   ```bash
   kubectl get pods --all-namespaces --sort-by='.status.containerStatuses[0].restartCount'
   # or from the alert labels: namespace + pod
   kubectl get pod <pod-name> -n <namespace>
   ```

2. **Check exit reason**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   # Look for: Last State, Reason (OOMKilled, Error, etc.)
   ```

3. **Check current and previous logs**
   ```bash
   kubectl logs <pod-name> -n <namespace> --tail=200
   kubectl logs <pod-name> -n <namespace> --previous
   ```

4. **Check resource limits**
   ```bash
   kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.spec.containers[*].resources}'
   ```

5. **Check events**
   ```bash
   kubectl get events -n <namespace> --sort-by='.lastTimestamp' | grep <pod-name>
   ```

6. **Check liveness/readiness probe config**
   ```bash
   kubectl describe pod <pod-name> -n <namespace> | grep -A 10 "Liveness\|Readiness"
   ```

## Remediation

### OOMKilled — increase memory limit
```bash
kubectl edit deployment/<deployment> -n <namespace>
# Increase resources.limits.memory
```

### Application crash — fix the code or config
```bash
# Check ConfigMaps / Secrets the pod depends on
kubectl get configmap -n <namespace>
kubectl get secret -n <namespace>
```

### Probe misconfigured
```bash
kubectl edit deployment/<deployment> -n <namespace>
# Adjust initialDelaySeconds, periodSeconds, failureThreshold
```

### Temporary relief (delete pod to force clean restart)
```bash
kubectl delete pod <pod-name> -n <namespace>
```

### Verify recovery
```bash
kubectl get pod <pod-name> -n <namespace> -w
# Watch RESTARTS column stabilize at 0 new restarts
```

## Escalation Path

1. **0–10 min**: On-call engineer diagnoses crash reason
2. **10–30 min**: If OOM or config issue, involve service owner
3. **30+ min**: If systemic (multiple pods), escalate to engineering lead

---
*Source: [observability/prometheus/rules/alerting.yml](../../observability/prometheus/rules/alerting.yml)*
