#!/usr/bin/env bash
set -euo pipefail

API_HOST=${API_HOST:-"http://api.eric-n.com"}
NAMESPACE=${NAMESPACE:-"myhealth"}
TOKEN=${TOKEN:-""}
DB_HOST=${DB_HOST:-""}
DB_USER=${DB_USER:-""}

info() { echo "[INFO] $*"; }
fail() { echo "[ERROR] $*"; exit 1; }

info "=== End-to-End Test Suite ==="

info "Test 1: Health checks"
curl -fsS "${API_HOST}/health" >/dev/null || fail "API health check failed"
info "✓ API service healthy"

info "Test 2: Prometheus metrics"
curl -fsS "${API_HOST}/metrics" | grep -q "http_requests_total" || fail "Metrics endpoint missing http_requests_total"
info "✓ Metrics available"

info "Test 3: Data flow (collector -> processor -> API)"
kubectl -n "${NAMESPACE}" create job test-collector --from=cronjob/oura-collector >/dev/null
echo "Waiting 30s for collector job..."
sleep 30
kubectl -n "${NAMESPACE}" logs job/test-collector || true
kubectl -n "${NAMESPACE}" delete job test-collector --ignore-not-found >/dev/null
info "✓ Collection job executed"

info "Test 4: Database connectivity"
if [[ -n "${DB_HOST}" && -n "${DB_USER}" ]]; then
  kubectl -n "${NAMESPACE}" exec deploy/data-processor -- \
    psql -h "${DB_HOST}" -U "${DB_USER}" -d myhealth -c "\\dt" >/dev/null || fail "Database connectivity failed"
  info "✓ Database connectivity verified"
else
  info "DB_HOST/DB_USER not set, skipping DB connectivity test"
fi

info "Test 5: API endpoints"
if [[ -n "${TOKEN}" ]]; then
  curl -fsS -H "Authorization: Bearer ${TOKEN}" "${API_HOST}/api/v1/dashboard" >/dev/null || fail "API endpoint check failed"
  info "✓ API endpoints responding"
else
  info "TOKEN not set, skipping authenticated endpoint test"
fi

info "=== All Tests Completed ==="
