# Runbook: Missing CA Trust Chain

## Symptoms
- Client cannot connect to protected service
- TLS handshake failure: "unknown certificate authority"
- mTLS connection rejected
- Trace shows `protected-access-error`

## Root Cause
- Client doesn't have CA root certificate
- CA chain is not properly installed on client
- Intermediate CA certificate is missing
- Certificate trust store is not configured

## Immediate Action
1. Verify CA service is up: `./scripts/run-healthcheck.sh`
2. Check protected service is reachable: `curl http://localhost:8004/health` (non-TLS port if available)
3. Review enrollment trace for certificate chain details

## Investigation Steps

### Step 1: Verify CA Root Certificate
```bash
# Get root certificate
curl http://localhost:8002/certs/root
```
Should return valid PEM block starting with `-----BEGIN CERTIFICATE-----`

### Step 2: Test Certificate Chain
```bash
# Get root cert
ROOT_CERT=$(curl -s http://localhost:8002/certs/root)

# Save to file
echo "$ROOT_CERT" > /tmp/ca-root.pem

# Verify it's valid
openssl x509 -in /tmp/ca-root.pem -text -noout
```

### Step 3: Check Certificate Request Response
```bash
# Make a cert signing request and check response includes CA chain
curl -s -X POST http://localhost:8002/certs/sign \
  -H "Content-Type: application/json" \
  -d '{"token":"test","client_cn":"test","validity_days":1}' | jq .ca_chain
```
Response should include both intermediate and root certificates

### Step 4: Review Enrollment Agent
```bash
docker compose logs enrollment-agent | grep -i "chain\|trust\|ca"
```

## Resolution

### Option A: Ensure Full CA Chain in Enrollment
1. Check `services/ca-service/cmd/main.go` in `signHandler()`
2. Verify response includes full CA chain:
   ```go
   caChain := caInstance.IntermediateCA.CertPEM + "\n" + caInstance.RootCA.CertPEM
   ```
3. Rebuild if needed: `docker compose up -d --build ca-service`

### Option B: Configure Client Trust Store
In enrollment agent, verify certificate chain is saved:
```bash
# Check enrollment response contains ca_chain
curl -s http://localhost:8003/enroll | jq .ca_chain
```

### Option C: Fix Certificate Generation
1. Check CA initialization: `services/ca-service/internal/ca.go`
2. Ensure both root and intermediate CAs are created
3. Verify intermediate is signed by root

## Verification
1. Get fresh certificate with full chain:
   ```bash
   RESPONSE=$(curl -s -X POST http://localhost:8002/certs/sign \
     -H "Content-Type: application/json" \
     -d '{"token":"test","client_cn":"test-cert","validity_days":1}')
   
   echo "$RESPONSE" | jq .ca_chain | wc -l
   # Should show multiple certificate blocks
   ```

2. Verify chain integrity:
   ```bash
   echo "$RESPONSE" | jq -r .ca_chain > chain.pem
   openssl verify -CAfile chain.pem chain.pem
   ```

3. Test enrollment via dashboard

## Prevention
- Ensure CA chain is always included in signing responses
- Validate chain completeness before sending to client
- Add tests for certificate chain
- Document expected chain depth (root + intermediate + client)

## Related Issues
- [missing-ca-trust-chain](missing-ca-trust-chain.md) - Client can't load CA cert into trust store
- [bad-client-cert](bad-client-cert.md) - Certificate was signed but is invalid
- [expired-oauth-token](expired-oauth-token.md) - Token invalid, preventing cert issuance

## Escalation
If issue persists:
1. Check CA logs: `docker compose logs ca-service --follow`
2. Inspect certificate chain generation: `services/ca-service/internal/ca.go`
3. Verify x509 package is correctly parsing certificates
4. Check for intermediate CA configuration issues
