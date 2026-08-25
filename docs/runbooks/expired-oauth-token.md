# Runbook: Expired OAuth Token

## Symptoms
- Enrollment fails at the token acquisition step
- Error message: "Token has expired" or "Token validation failed"
- Trace shows `token-failed` event

## Root Cause
- IdP token TTL is too short
- Token was issued but not used within expiration window
- System clock is skewed (token appears expired before it should be)
- Token refresh mechanism failed

## Immediate Action
1. Check IdP service is running: `./scripts/run-healthcheck.sh`
2. View the failed enrollment trace in the dashboard
3. Note the trace ID for incident creation

## Investigation Steps

### Step 1: Verify IdP Service
```bash
curl http://localhost:8001/health
```
Expected response: `{"status":"ok","service":"idp-mock"}`

### Step 2: Check Token Generation
```bash
# Request a token directly
curl -X POST http://localhost:8001/token \
  -H "Content-Type: application/json" \
  -d '{"grant_type":"password","username":"device-123","password":"pass","audience":"ca-service"}'
```
Look for `expires_in` field - default should be 3600 seconds (1 hour)

### Step 3: Review IdP Configuration
- Check `services/idp-mock/cmd/main.go` for token TTL
- Default: 3600 seconds
- Can be adjusted in `GenerateToken()` call

### Step 4: Check System Time
```bash
# On each container
docker exec stepdeploy-lab-idp-mock date
docker exec stepdeploy-lab-ca-service date
docker exec stepdeploy-lab-enrollment-agent date
```
Times should be within 1 minute of each other

## Resolution

### Option A: Increase Token TTL
1. Edit `services/idp-mock/cmd/main.go`
2. Change token expiration in `tokenHandler()`:
   ```go
   token, err := auth.GenerateToken(tokenReq.Username, tokenReq.Audience, 7200) // 2 hours
   ```
3. Rebuild and restart IdP service:
   ```bash
   docker compose up -d --build idp-mock
   ```

### Option B: Fix Clock Skew
1. Resynchronize container clocks:
   ```bash
   docker compose down
   docker system prune -a
   docker compose up -d
   ```

### Option C: Fast-Track for Demo
- Reset injected failures: `./scripts/reset-environment.sh`
- Run enrollment again: Dashboard → Test → Run Enrollment Test

## Verification
1. Run health check: `./scripts/run-healthcheck.sh`
2. Test enrollment via dashboard
3. Verify trace shows `token-ok` event
4. Check incident SLA is within target

## Prevention
- Monitor token TTL in production (should be 1-4 hours)
- Implement token refresh before expiration
- Set up alerts for clock skew
- Add token expiration warnings to logs

## Escalation
If issue persists after these steps:
1. Check IdP logs: `docker compose logs idp-mock`
2. Check CA logs: `docker compose logs ca-service`
3. Collect logs: `./scripts/collect-logs.sh`
4. Review `/diagnostics/traces` for detailed event timeline
