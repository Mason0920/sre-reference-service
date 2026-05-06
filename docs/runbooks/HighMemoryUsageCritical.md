# Runbook: HighMemoryUsageCritical

| Field | Value |
|-------|-------|
| **Severity** | critical |
| **Source** | `observability/prometheus/rules/alerting.yml` |
| **Alert Group** | `alerts.infrastructure` |
| **Condition** | Memory usage > 95% for 5m |
| **Slack Channel** | `#alerts` |
| **Related Alert** | [HighMemoryUsage](./HighMemoryUsage.md) (>85%) |

## Summary

Node memory usage has exceeded 95% for 5 minutes. The kernel OOM killer may start terminating processes. **Act immediately.**

## Symptoms

- Slack alert: `[FIRING] HighMemoryUsageCritical`
- Pods being OOMKilled (`kubectl describe pod` shows `OOMKilled`)
- System processes becoming unresponsive
- `dmesg` shows `Out of memory: Kill process`

## Root Cause Analysis

1. **Immediate: check available memory**
   ```bash
   free -h
   # If available < 500MB, treat as emergency
   ```

2. **Check OOM kills**
   ```bash
   dmesg | grep -i "out of memory" | tail -20
   journalctl -k | grep -i "oom" | tail -20
   ```

3. **Find top memory consumers**
   ```bash
   ps aux --sort=-%mem | head -15
   kubectl top pods --all-namespaces --sort-by=memory | head -15
   ```

4. **Check if swap is exhausted**
   ```bash
   swapon -s
   cat /proc/swaps
   ```

5. **Check pod memory limits**
   ```bash
   kubectl get pod -n <namespace> -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].resources.limits.memory}{"\n"}{end}'
   ```

## Remediation (by priority)

### 1. Free memory immediately — restart highest consumer
```bash
# Identify and restart the largest pod
kubectl rollout restart deployment/<top-memory-deployment> -n <namespace>
# Verify
kubectl top pods -n <namespace>
```

### 2. Drop caches (temporary, Linux)
```bash
sync && echo 1 > /proc/sys/vm/drop_caches
free -h
```

### 3. Evict low-priority pods
```bash
kubectl drain <node> --ignore-daemonsets --delete-emptydir-data
```

### 4. Add node to cluster (if cloud)
```bash
# AWS: increase ASG desired count
# GKE: resize node pool
```

### 5. Set memory limits on pods without them
```bash
kubectl edit deployment/<deployment> -n <namespace>
# Add: resources.limits.memory: "512Mi"
```

### Verify recovery
```bash
free -h
kubectl top nodes
# Target: < 85% before resolving alert
```

## Escalation Path

1. **Immediate**: On-call restarts top consumer, checks OOM kills
2. **< 5 min**: If multiple pods affected, wake infra lead
3. **If services down**: Trigger incident, notify engineering lead

---
*Source: [observability/prometheus/rules/alerting.yml](../../observability/prometheus/rules/alerting.yml)*
