#!/bin/bash

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <failure_type>"
    echo ""
    echo "Available failure types:"
    echo "  expired-oauth-token    - IdP returns expired token"
    echo "  wrong-oauth-audience   - IdP returns token with wrong audience"
    echo "  missing-ca-trust       - Client doesn't trust CA cert"
    echo "  bad-client-cert        - CA refuses to sign cert"
    echo "  service-unavailable    - Protected service returns 503"
    echo "  dns-mismatch           - DNS config mismatch"
    echo "  clock-skew             - Time sync issues"
    exit 1
fi

FAILURE_TYPE=$1

echo "=== Injecting Failure: $FAILURE_TYPE ==="
echo ""

# Service URLs
IDP_URL="${IDP_URL:-http://localhost:8001}"
CA_URL="${CA_URL:-http://localhost:8002}"
ENROLLMENT_AGENT_URL="${ENROLLMENT_AGENT_URL:-http://localhost:8003}"
PROTECTED_SERVICE_URL="${PROTECTED_SERVICE_URL:-https://localhost:8004}"

# Inject failure into appropriate service(s)
case $FAILURE_TYPE in
    expired-oauth-token|wrong-oauth-audience)
        echo "Injecting into IdP Mock..."
        curl -s -X POST "$IDP_URL/test/inject-failure" \
            -H "Content-Type: application/json" \
            -d "{\"failure_type\":\"$FAILURE_TYPE\"}" | jq .
        ;;
    bad-client-cert)
        echo "Injecting into CA Service..."
        curl -s -X POST "$CA_URL/test/inject-failure" \
            -H "Content-Type: application/json" \
            -d "{\"failure_type\":\"$FAILURE_TYPE\"}" | jq .
        ;;
    missing-ca-trust|dns-mismatch|clock-skew)
        echo "Injecting into Enrollment Agent..."
        curl -s -X POST "$ENROLLMENT_AGENT_URL/test/inject-failure" \
            -H "Content-Type: application/json" \
            -d "{\"failure_type\":\"$FAILURE_TYPE\"}" | jq .
        ;;
    service-unavailable)
        echo "Injecting into Protected Service..."
        curl -s -k -X POST "$PROTECTED_SERVICE_URL/test/inject-failure" \
            -H "Content-Type: application/json" \
            -d "{\"failure_type\":\"$FAILURE_TYPE\"}" | jq .
        ;;
    *)
        echo "Unknown failure type: $FAILURE_TYPE"
        exit 1
        ;;
esac

echo ""
echo "Failure injected successfully!"
echo "Run 'docker compose logs -f' to see effects"
echo ""
echo "To reset all failures, run: ./scripts/reset-environment.sh"
