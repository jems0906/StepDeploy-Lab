# Runbook: Bad Client Certificate

## Symptoms
- Enrollment fails at certificate validation step
- Error: "Certificate signing failed" or "Certificate signature verification failed"
- Protected service returns 403 Forbidden
- Trace shows `cert-request-error` or `protected-access-error`

## Root Cause
- Client certificate is expired
- Certificate was not properly signed by CA
- Certificate doesn't match client identity
- Certificate chain is incomplete
- Public key doesn't match private key

## Immediate Action
1. Check CA service is running: `./scripts/run-healthcheck.sh`
2. Review the failed trace in dashboard
3. Note certificate fingerprint if available in logs

## Investigation Steps

### Step 1: Verify CA Service
```bash
curl http://localhost:8002/health
```
Expected: `{"status":"ok","service":"ca-service","issued_certs": X}`

### Step 2: Get CA Root Certificate
```bash
curl http://localhost:8002/certs/root
```
Should return valid PEM-formatted root certificate

### Step 3: Review CA Logs
```bash
docker compose logs ca-service | tail -20
```
Look for errors like:
- "Certificate issuance failed"
- "Token validation failed"
- "Signature verification failed"

### Step 4: Check Certificate Validity
```bash
# Get a certificate manually
CERT=$(curl -s -X POST http://localhost:8002/certs/sign \
  -H "Content-Type: application/json" \
  -d '{"token":"valid-token","client_cn":"test-device","validity_days":1}' | jq -r .certificate)

# Check certificate details
echo "$CERT" | openssl x509 -text -noout
```

## Resolution

### Option A: Renew Certificate
1. Reset enrollment state: `./scripts/reset-environment.sh`
2. Trigger new enrollment via dashboard
3. CA will issue fresh certificate

### Option B: Extend Certificate Validity
1. Edit `services/ca-service/internal/provisioner.go`
2. Change `validityDays` parameter (default: 1 day)
3. Rebuild:
   ```bash
   docker compose up -d --build ca-service
   ```

### Option C: Check Certificate Chain
1. Verify certificate is signed by CA root:
   ```bash
   # Get client cert and CA cert
   CLIENT_CERT=$(...)  # From enrollment response
   CA_CERT=$(curl http://localhost:8002/certs/root)
   
   # Verify signature
   echo "$CA_CERT" > ca.pem
   echo "$CLIENT_CERT" > client.pem
   openssl verify -CAfile ca.pem client.pem
   ```

## Verification
1. Run health check: `./scripts/run-healthcheck.sh`
2. Test enrollment via dashboard
3. Verify trace shows `cert-ok` event
4. Check protected service allows access with new cert

## Prevention
- Monitor certificate expiration dates
- Implement certificate auto-renewal 30 days before expiry
- Add cert validity checks in enrollment workflow
- Alert on short-lived certificates

## Escalation
If issue persists:
1. Check CA container logs: `docker compose logs ca-service --follow`
2. Inspect provisioner code: `services/ca-service/internal/provisioner.go`
3. Verify root CA is valid: `curl http://localhost:8002/certs/root | openssl x509 -text`
4. Collect full logs: `./scripts/collect-logs.sh`
