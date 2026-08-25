#!/bin/bash

set -e

echo "=== Resetting Environment ==="
echo ""

# Service URLs
IDP_URL="${IDP_URL:-http://localhost:8001}"
CA_URL="${CA_URL:-http://localhost:8002}"
ENROLLMENT_AGENT_URL="${ENROLLMENT_AGENT_URL:-http://localhost:8003}"
PROTECTED_SERVICE_URL="${PROTECTED_SERVICE_URL:-https://localhost:8004}"

# Reset all services
echo "Resetting failures in all services..."
curl -s -X POST "$IDP_URL/test/reset-failures" -H "Content-Type: application/json" || true
curl -s -X POST "$CA_URL/test/reset-failures" -H "Content-Type: application/json" || true
curl -s -X POST "$ENROLLMENT_AGENT_URL/test/reset-failures" -H "Content-Type: application/json" || true
curl -s -k -X POST "$PROTECTED_SERVICE_URL/test/reset-failures" -H "Content-Type: application/json" || true

echo ""
echo "Stopping Docker containers..."
docker compose stop

echo "Removing containers..."
docker compose rm -f

echo ""
echo "Starting fresh environment..."
docker compose up -d

echo ""
echo "Waiting for services to start..."
sleep 5

echo "Running health check..."
./scripts/run-healthcheck.sh

echo ""
echo "=== Environment Reset Complete ==="
