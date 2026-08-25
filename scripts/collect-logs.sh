#!/bin/bash

set -e

echo "=== Collecting Logs ==="
echo ""

OUTPUT_DIR="logs-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUTPUT_DIR"

echo "Collecting logs to $OUTPUT_DIR/"

docker compose logs idp-mock > "$OUTPUT_DIR/idp-mock.log"
docker compose logs ca-service > "$OUTPUT_DIR/ca-service.log"
docker compose logs enrollment-agent > "$OUTPUT_DIR/enrollment-agent.log"
docker compose logs protected-service > "$OUTPUT_DIR/protected-service.log"
docker compose logs diagnostics-api > "$OUTPUT_DIR/diagnostics-api.log"
docker compose logs frontend > "$OUTPUT_DIR/frontend.log"

echo "Collecting service stats..."
docker compose stats --no-stream > "$OUTPUT_DIR/docker-stats.txt"

echo "Collecting diagnostics info..."
curl -s http://localhost:8080/dependencies > "$OUTPUT_DIR/dependencies.json" || true
curl -s http://localhost:8080/traces > "$OUTPUT_DIR/traces.json" || true
curl -s http://localhost:8080/incidents > "$OUTPUT_DIR/incidents.json" || true

echo ""
echo "Logs collected in $OUTPUT_DIR"
echo ""
echo "Log files:"
ls -lh "$OUTPUT_DIR/"
