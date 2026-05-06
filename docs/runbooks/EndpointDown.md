# Runbook: EndpointDown

| Field | Value |
|-------|-------|
| **Severity** | critical |
| **Source** | `observability/prometheus/rules/slo.yml` |
| **Alert Group** | `alerts.availability` |
| **Condition** | `probe_success == 0` for 1m |
| **SLO Impact** | Full availability SLO breach |
| **Slack Channel** | `#alerts` |

## Summary

A blackbox exporter probe has reported `probe_success == 0` for 1 minute. The endpoint is completely unreachable from the monitoring network.

## Symptoms

- Slack alert: `[FIRING] EndpointDown` with `instance` label
- Prometheus shows `probe_success{instance="<url>"} == 0`
- External monitoring / synthetic checks failing
- Users unable to access the service

## Root Cause Analysis

1. **Confirm the probe is failing**
   ```bash
   curl -s "http://localhost:9115/probe?target=<instance>&module=http_2xx" | grep probe_success
   # Should return probe_success 0 if alert is firing
   ```

2. **Test direct connectivity**
   ```bash
   curl -v --max-time 10 <instance>
   # Check: DNS resolution, TCP connection, TLS, HTTP response code
   ```

3. **Check DNS resolution**
   ```bash
   nslookup <hostname>
   dig <hostname>
   ```

4. **Check if service pods are running**
   ```bash
   kubectl get pods -n <namespace> -l app=<service>
   kubectl get endpoints <service> -n <namespace>
   ```

5. **Check ingress / load balancer**
   ```bash
   kubectl get ingress -n <namespace>
   kubectl describe ingress <ingress-name> -n <namespace>
   kubectl get svc -n <namespace>
   ```

6. **Check TLS certificate (if HTTPS)**
   ```bash
   echo | openssl s_client -connect <hostname>:443 2>/dev/null | openssl x509 -noout -dates
   ```

7. **Check network policies**
   ```bash
   kubectl get networkpolicy -n <namespace>
   ```

## Remediation

### Service pods down — restart
```bash
kubectl rollout restart deployment/<service> -n <namespace>
kubectl rollout status deployment/<service> -n <namespace>
```

### Ingress not routing — check annotations
```bash
kubectl describe ingress <name> -n <namespace>
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller | tail -50
```

### Expired TLS certificate
```bash
kubectl get certificate -n <namespace>  # cert-manager
kubectl describe certificate <cert-name> -n <namespace>
kubectl delete certificate <cert-name> -n <namespace>  # triggers renewal
```

### DNS misconfiguration
```bash
kubectl get svc <service> -n <namespace>  # check ClusterIP / external IP
# Update DNS record to point to correct IP
```

### Verify recovery
```bash
curl -s "http://localhost:9115/probe?target=<instance>&module=http_2xx" | grep probe_success
# Should return probe_success 1
```

## Escalation Path

1. **Immediate**: On-call verifies scope (1 endpoint vs all services)
2. **< 5 min**: If full service outage, trigger incident immediately
3. **< 5 min**: Notify engineering lead; consider customer status page update
4. **If DNS/cert/infra issue**: Escalate to infra team immediately

---
*Source: [observability/prometheus/rules/slo.yml](../../observability/prometheus/rules/slo.yml)*
