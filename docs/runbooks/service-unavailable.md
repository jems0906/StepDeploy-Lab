# Runbook: Service Unavailable

## Symptoms
- Enrollment fails with "service unavailable" error
- Specific service returns 503 Service Unavailable
- Service containers show as "down" in dependency map
- Trace shows step failure without timeout

## Root Cause
- Service container crashed or stopped
- Service failed to start on boot
- Out of memory (OOM) error
- Port conflict or bind error
- Service dependency not ready

## Immediate Action
1. Check which service is down: `./scripts/run-healthcheck.sh`
2. View service status: `docker compose ps`
3. Check service logs: `docker compose logs <service-name>`

## Investigation Steps

### Step 1: Check Service Status
```bash
docker compose ps
```
Look for services not in "Up" state. Status column shows:
- `Up X seconds` - Service is running
- `Exited (0)` - Service stopped normally
- `Exited (1)` - Service crashed
- `Exit (137)` - Out of memory kill

### Step 2: View Service Logs
```bash
# Check specific service
docker compose logs ca-service --tail 50

# Follow logs in real-time
docker compose logs -f protected-service
```
Look for:
- Port already in use
- Connection refused errors
- "listen: bind: address already in use"
- Panic messages
- Initialization errors

### Step 3: Check Resource Constraints
```bash
# View container resource usage
docker compose stats

# Check memory limits
docker compose inspect ca-service | jq '.[0].HostConfig.Memory'
```

### Step 4: Verify Port Availability
```bash
# Check if ports are in use
netstat -an | grep LISTEN | grep 800
```
Expected ports:
- 8001 - IdP Mock
- 8002 - CA Service
- 8003 - Enrollment Agent
- 8004 - Protected Service
- 8080 - Diagnostics API
- 5173 - Frontend

## Resolution

### Option A: Restart Service
```bash
# Restart specific service
docker compose restart ca-service

# Wait for it to start
sleep 3

# Verify health
curl http://localhost:8002/health
```

### Option B: Rebuild and Restart
```bash
# Rebuild service image
docker compose up -d --build ca-service

# Monitor startup
docker compose logs -f ca-service
```

### Option C: Full Environment Reset
```bash
./scripts/reset-environment.sh
```
This will:
1. Stop all containers
2. Remove containers and networks
3. Start fresh environment
4. Run health checks

### Option D: Check Service Dependencies
```bash
# Verify IdP is ready before CA starts
docker compose logs ca-service | grep -i "idp\|connect"

# Check if service is waiting for another service
docker compose logs enrollment-agent | grep -i "timeout\|refused"
```

### Option E: Increase Resource Limits
Edit `docker-compose.yml`:
```yaml
services:
  ca-service:
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```
Then restart: `docker compose up -d`

## Verification
1. Check all services up: `./scripts/run-healthcheck.sh`
2. Monitor for 30 seconds:
   ```bash
   watch -n 1 'docker compose ps'
   ```
3. Test enrollment via dashboard
4. Verify no errors in logs:
   ```bash
   docker compose logs | grep -i "error\|fatal\|panic"
   ```

## Prevention
- Set restart policy in docker-compose.yml:
  ```yaml
  restart_policy:
    condition: on-failure
    max_attempts: 3
  ```
- Monitor container health with health checks
- Set appropriate resource limits
- Log rotate to prevent disk full
- Monitor free disk space

## Common Issues

### Port Already in Use
```bash
# Find what's using the port
lsof -i :8002

# Kill the process
kill -9 <PID>
```

### Out of Memory
```bash
# Check available memory
free -h

# Check container memory usage
docker stats --no-stream
```
**Fix**: Increase system memory or reduce container memory overhead

### Dependency Not Ready
```bash
# Check service startup order
docker-compose config | grep depends_on
```
**Fix**: Verify health check conditions and startup order

## Related Runbooks
- [expired-oauth-token](expired-oauth-token.md) - IdP service degradation
- [bad-client-cert](bad-client-cert.md) - CA service not responding
- [service-unavailable](service-unavailable.md) - Protected service down

## Escalation
If issue persists:
1. Collect detailed logs: `./scripts/collect-logs.sh`
2. Check Docker daemon logs: `docker logs --tail 50`
3. Inspect service configuration: `docker compose config`
4. Verify image integrity: `docker images | grep stepdeploy`
5. Try full cleanup and rebuild:
   ```bash
   docker compose down -v
   docker system prune -a
   docker compose up -d --build
   ```
