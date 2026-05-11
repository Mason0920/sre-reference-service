# Runbook: HighMemoryUsage

| Field | Value |
|-------|-------|
| **Severity** | warning |
| **Source** | `observability/prometheus/rules/alerting.yml` |
| **Alert Group** | `alerts.infrastructure` |
| **Condition** | Memory usage > 85% for 5m |
| **Slack Channel** | `#alerts` |
| **Related Alert** | [HighMemoryUsageCritical](./HighMemoryUsageCritical.md) (>95%) |

## Summary

Node memory usage has been above 85% for at least 5 minutes. The system may begin swapping or OOM-killing processes if usage continues to rise.

## Symptoms

- Slack alert: `[FIRING] HighMemoryUsage`
- `free -h` shows high memory consumption
- Possible increase in pod restart counts (OOMKill precursor)

## Root Cause Analysis

1. **Check current memory on the node**
   ```bash
   free -h
   cat /proc/meminfo | head -20
   ```

2. **Identify top memory-consuming processes**
   ```bash
   ps aux --sort=-%mem | head -20
   # or
   top -o %MEM -b -n 1 | head -30
   ```

3. **Check Kubernetes pod memory consumption**
   ```bash
   kubectl top pods --all-namespaces --sort-by=memory | head -20
   kubectl top nodes
   ```

4. **Check Prometheus query for trend**
   ```
   (node_memory_MemTotal_bytes{instance="<instance>"} - node_memory_MemAvailable_bytes{instance="<instance>"}) / node_memory_MemTotal_bytes{instance="<instance>"}
   ```

5. **Check for memory leaks (growth over time)**
   ```
   # In Prometheus: check 1h trend
   delta((node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes)[1h:5m])
   ```

6. **Check pod memory limits vs usage**
   ```bash
   kubectl get pod <pod> -n <namespace> -o jsonpath='{.spec.containers[*].resources}'
   kubectl top pod <pod> -n <namespace>
   ```

## Remediation

### Restart high-memory pod (short-term)
```bash
kubectl rollout restart deployment/<deployment> -n <namespace>
```

### Increase pod memory limit
```bash
kubectl edit deployment/<deployment> -n <namespace>
# Increase resources.limits.memory and resources.requests.memory
```

### Clear Linux page cache (careful — temporary)
```bash
sync && echo 3 > /proc/sys/vm/drop_caches
```

### Scale horizontally if traffic-driven
```bash
kubectl scale deployment/<deployment> --replicas=<n> -n <namespace>
```

### Verify recovery
```bash
free -h
kubectl top nodes
```

## Escalation Path

1. **0–15 min**: On-call identifies top consumers and restarts if needed
2. **15–30 min**: Involve service owner if memory leak suspected
3. **If > 90%**: Treat as critical — see [HighMemoryUsageCritical](./HighMemoryUsageCritical.md)

---
*Source: [observability/prometheus/rules/alerting.yml](../../observability/prometheus/rules/alerting.yml)*
