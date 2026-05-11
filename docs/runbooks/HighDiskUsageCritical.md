# Runbook: HighDiskUsageCritical

| Field | Value |
|-------|-------|
| **Severity** | critical |
| **Source** | `observability/prometheus/rules/alerting.yml` |
| **Alert Group** | `alerts.infrastructure` |
| **Condition** | Disk usage > 95% for 5m |
| **Slack Channel** | `#alerts` |
| **Related Alert** | [HighDiskUsage](./HighDiskUsage.md) (>85%) |

## Summary

Disk usage has exceeded 95%. Services writing to disk (databases, log collectors, Prometheus) will begin failing imminently. **Act immediately.**

## Symptoms

- Slack alert: `[FIRING] HighDiskUsageCritical`
- Write errors in application logs: `no space left on device`
- Database write failures, Prometheus write errors
- Pod evictions if `/var/lib/kubelet` is affected

## Root Cause Analysis

1. **Check remaining space**
   ```bash
   df -h <mountpoint>
   # If < 1GB free, treat as emergency
   ```

2. **Find what's consuming space** (fast scan)
   ```bash
   du -xsh /* 2>/dev/null | sort -rh | head -15
   du -xsh /var/log/* 2>/dev/null | sort -rh | head -10
   du -xsh /tmp/* 2>/dev/null | sort -rh | head -10
   ```

3. **Check for deleted-but-open files** (common with logs)
   ```bash
   lsof +L1 2>/dev/null | awk '{print $7, $9}' | sort -rn | head -10
   ```

4. **Check Prometheus TSDB size**
   ```bash
   du -sh /path/to/prometheus/data
   ```

5. **Check Docker**
   ```bash
   docker system df
   ```

## Remediation (by priority — fastest first)

### 1. Free space immediately
```bash
# Remove old journal logs
journalctl --vacuum-size=500M

# Remove rotated/compressed logs
find /var/log -name "*.gz" -delete
find /var/log -name "*.1" -delete

# Truncate large active log (do NOT delete)
> /var/log/some-large.log
```

### 2. Free deleted-but-open file space
```bash
# Identify PID holding deleted file, then restart that process
kill -HUP <pid>
```

### 3. Docker cleanup
```bash
docker system prune -af --volumes
```

### 4. Expand volume (immediate, cloud)
```bash
# AWS: resize EBS volume in console, then:
sudo growpart /dev/xvda 1
sudo xfs_growfs /
# Verify
df -h
```

### 5. Reduce Prometheus retention (emergency)
```bash
# Edit observability/docker-compose.yml:
# command: --storage.tsdb.retention.time=7d
docker compose -f observability/docker-compose.yml up -d prometheus
```

### Verify recovery
```bash
df -h <mountpoint>
# Target: < 85% before declaring resolved
```

## Escalation Path

1. **Immediate**: On-call engineer takes emergency remediation steps above
2. **< 5 min**: If cannot free space quickly, wake infra lead for volume expansion
3. **If services affected**: Trigger incident, notify engineering lead and service owners

---
*Source: [observability/prometheus/rules/alerting.yml](../../observability/prometheus/rules/alerting.yml)*
