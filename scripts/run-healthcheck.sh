#!/bin/bash

set -e

echo "=== StepDeploy Lab Health Check ==="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service URLs
IDP_URL="${IDP_URL:-http://localhost:8001}"
CA_URL="${CA_URL:-http://localhost:8002}"
ENROLLMENT_AGENT_URL="${ENROLLMENT_AGENT_URL:-http://localhost:8003}"
PROTECTED_SERVICE_URL="${PROTECTED_SERVICE_URL:-https://localhost:8004}"
DIAGNOSTICS_URL="${DIAGNOSTICS_URL:-http://localhost:8080}"

# Function to check service health
check_service() {
    local name=$1
    local url=$2
    
    if curl -s -k "$url/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $name: UP"
        return 0
    else
        echo -e "${RED}✗${NC} $name: DOWN"
        return 1
    fi
}

# Check all services
echo "Checking services..."
all_up=0
check_service "IdP Mock" "$IDP_URL" || all_up=1
check_service "CA Service" "$CA_URL" || all_up=1
check_service "Enrollment Agent" "$ENROLLMENT_AGENT_URL" || all_up=1
check_service "Protected Service" "$PROTECTED_SERVICE_URL" || all_up=1
check_service "Diagnostics API" "$DIAGNOSTICS_URL" || all_up=1

echo ""
if [ $all_up -eq 0 ]; then
    echo -e "${GREEN}All services are healthy!${NC}"
else
    echo -e "${RED}Some services are down.${NC}"
    exit 1
fi

# Check dependencies from diagnostics API
echo ""
echo "Checking dependency graph..."
curl -s -k "$DIAGNOSTICS_URL/dependencies" | jq -r '.services[] | "\(.name): \(.status)"'

echo ""
echo "=== Health Check Complete ==="
