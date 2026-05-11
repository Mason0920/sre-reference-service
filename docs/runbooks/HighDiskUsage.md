# Runbook: HighDiskUsage

| Field | Value |
|-------|-------|
| **Severity** | warning |
| **Source** | `observability/prometheus/rules/alerting.yml` |
| **Alert Group** | `alerts.infrastructure` |
| **Condition** | Disk usage > 85% for 5m |
| **Slack Channel** | `#alerts` |
| **Related Alert** | [HighDiskUsageCritical](./HighDiskUsageCritical.md) (>95%) |

## Summary

Disk usage on a node has exceeded 85% for at least 5 minutes. Action is needed before the disk fills completely and causes service failures.

## Symptoms

- Slack alert: `[FIRING] HighDiskUsage` with `instance` and `mountpoint` labels
- `df -h` shows ≥85% on the affected mount
- Log writes or database operations may slow down

## Root Cause Analysis

1. **Check current disk usage**
   ```bash
   df -h <mountpoint>
   du -sh /* 2>/dev/null | sort -rh | head -20
   ```

2. **Identify largest directories**
   ```bash
   du -sh /var/log/* 2>/dev/null | sort -rh | head -10
   du -sh /var/lib/docker/* 2>/dev/null | sort -rh | head -10
   ```

3. **Check Prometheus query for trend**
   ```
   # In Prometheus UI
   (node_filesystem_size_bytes{instance="<instance>",mountpoint="<mp>"} - node_filesystem_avail_bytes{instance="<instance>",mountpoint="<mp>"}) / node_filesystem_size_bytes{instance="<instance>",mountpoint="<mp>"}
   ```

4. **Check for large log files**
   ```bash
   find /var/log -name "*.log" -size +500M 2>/dev/null
   ```

5. **Check Docker disk usage (if applicable)**
   ```bash
   docker system df
   ```

## Remediation

### Clean up old logs
```bash
# Truncate large log files (do NOT delete active ones)
journalctl --vacuum-time=7d
find /var/log -name "*.gz" -mtime +30 -delete
```

### Clean Docker artifacts
```bash
docker system prune -f
docker image prune -a -f  # Remove unused images
docker volume prune -f
```

### Clean up Prometheus data (if filling /data)
```bash
# Reduce retention if needed — edit docker-compose.yml
# --storage.tsdb.retention.time=15d (default 15d)
```

### Expand volume (if cloud)
```bash
# AWS EBS — resize via console then:
sudo growpart /dev/xvda 1
sudo xfs_growfs /  # or resize2fs for ext4
```

### Verify recovery
```bash
df -h <mountpoint>
# Should show < 80% after cleanup
```

## Escalation Path

1. **0–15 min**: On-call engineer frees disk space
2. **15–30 min**: If unable to free enough space, escalate to infra team for volume expansion
3. **If > 90%**: Treat as critical — see [HighDiskUsageCritical](./HighDiskUsageCritical.md)

---
*Source: [observability/prometheus/rules/alerting.yml](../../observability/prometheus/rules/alerting.yml)*
