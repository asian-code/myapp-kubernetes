# Operations Runbook

## Accessing the Cluster

```bash
aws eks update-kubeconfig \
  --name myhealth \
  --region us-east-1 \
  --alias myhealth-dev

kubectl --context myhealth-dev get nodes
```

## Monitoring

```bash
# Grafana (port-forward)
kubectl -n myhealth port-forward svc/grafana 3000:3000
# http://localhost:3000 (admin / password from secret)

# Prometheus targets
kubectl -n myhealth port-forward svc/prometheus-server 9090:80
# http://localhost:9090/targets
```

## Troubleshooting

### Pod Stuck in Pending
```bash
kubectl describe pod <pod> -n myhealth | tail -50
# Check resource requests vs node capacity
# Check if PersistentVolumeClaims are bound
```

### Database Connection Issues
```bash
kubectl -n myhealth run -it --rm debug \
  --image=postgres:14 \
  --restart=Never -- \
  psql -h <RDS_ENDPOINT> -U myhealth_user -d myhealth
```

### API Gateway / DNS Issues
```bash
# Verify Route53 record
aws route53 list-resource-record-sets \
  --hosted-zone-id <ZONE_ID> \
  --query "ResourceRecordSets[?Name == 'api.eric-n.com.']"

# Verify API Gateway domain
aws apigatewayv2 get-domain-names --region us-east-1
```

## Backup & Restore

### Backup Database Snapshot
```bash
aws rds create-db-snapshot \
  --db-instance-identifier myhealth-db \
  --db-snapshot-identifier myhealth-db-backup-$(date +%Y%m%d)
```

### Restore From Snapshot
```bash
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier myhealth-db-restore \
  --db-snapshot-identifier <SNAPSHOT_ID>
```

## On-Call Quick Checks
- Prometheus targets are UP
- Grafana dashboards load
- AlertManager has no critical firing alerts
- API /health returns 200
- /metrics returns http_requests_total
- Pods: `kubectl get pods -n myhealth`
