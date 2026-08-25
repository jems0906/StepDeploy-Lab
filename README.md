# StepDeploy Lab

A multi-service Go demo that simulates a customer identity deployment, injects failures across services, and provides root-cause diagnostics with on-call/SLA-style incident tracking.

## Overview

This project demonstrates a Smallstep-style customer deployment with:
- OAuth/OIDC identity provider
- Certificate authority for issuing short-lived X.509 client certificates
- Enrollment agent for device-side operations
- Protected mTLS-enabled service
- Comprehensive diagnostics API and dashboard
- Failure injection for chaos engineering
- Incident tracking with SLA management

## Quick Start

### Local Development with Docker Compose

```bash
# Build and start all services
docker compose up -d

# Check health of all services
./scripts/run-healthcheck.sh

# Inject a failure (e.g., expired OAuth token)
./scripts/inject-failure.sh expired-oauth-token

# View logs
./scripts/collect-logs.sh

# Reset environment
./scripts/reset-environment.sh

# Access the dashboard
open http://localhost:5173
```

### Manual Development

#### Prerequisites
- Go 1.22+
- Node.js 18+
- Docker and Docker Compose

#### Setup

1. **Start services individually** (in separate terminals):

```bash
# IdP Mock
cd services/idp-mock
go run ./cmd/main.go

# CA Service
cd services/ca-service
go run ./cmd/main.go

# Protected Service
cd services/protected-service
go run ./cmd/main.go

# Diagnostics API
cd services/diagnostics-api
go run ./cmd/main.go

# Frontend
cd frontend
npm install
npm run dev
```

2. **Run health checks**:
```bash
./scripts/run-healthcheck.sh
```

3. **Inject failures**:
```bash
./scripts/inject-failure.sh expired-oauth-token
./scripts/inject-failure.sh missing-ca-trust
./scripts/inject-failure.sh bad-client-cert
./scripts/inject-failure.sh service-unavailable
./scripts/inject-failure.sh dns-mismatch
./scripts/inject-failure.sh clock-skew
```

## Architecture

### Services

1. **IdP Mock** (`services/idp-mock`)
   - Issues OAuth/OIDC tokens
   - Simulates Okta/Google Workspace behavior
   - Endpoints: `/token`, `/userinfo`

2. **CA Service** (`services/ca-service`)
   - Issues short-lived X.509 client certificates
   - Validates OAuth tokens
   - Manages root and intermediate CA
   - Endpoints: `/certs/sign`

3. **Enrollment Agent** (`services/enrollment-agent`)
   - Device-side enrollment workflow
   - Gets OAuth token from IdP
   - Requests client certificate from CA
   - Connects to protected service with mTLS

4. **Protected Service** (`services/protected-service`)
   - Requires valid client certificate via mTLS
   - Returns protected resources
   - Endpoint: `/api/resource`

5. **Diagnostics API** (`services/diagnostics-api`)
   - Health checks for all services
   - Trace collection and aggregation
   - Incident tracking with SLA management
   - Serves dashboard backend
   - Endpoints: `/health`, `/traces`, `/incidents`, `/dependencies`

### Frontend

React + Vite dashboard showing:
- Dependency health map
- Trace timeline with failure root causes
- Log aggregation by deployment step
- Incident tracking with severity and SLA timers
- Recommended runbooks for each failure

## Failure Injection Scenarios

The system supports injecting various failures:

- **expired-oauth-token**: Enrollment agent receives expired token
- **wrong-oauth-audience**: Token audience mismatch
- **missing-ca-trust**: Client doesn't trust CA root certificate
- **bad-client-cert**: Invalid or expired client certificate
- **service-unavailable**: Protected service offline
- **dns-mismatch**: DNS/config mismatch between services
- **clock-skew**: Time synchronization issues

See `docs/runbooks/` for detailed remediation guides.

## Testing

```bash
# Run all tests
go test ./...

# Run specific service tests
cd services/idp-mock && go test ./...
cd services/ca-service && go test ./...
```

## Linting

```bash
golangci-lint run ./...
```

## CI/CD

GitHub Actions workflow runs:
- Go tests across all services
- golangci-lint for code quality
- Docker image builds

## Deployment

### Local with Docker Compose
See [docker-compose.yml](docker-compose.yml)

### Railway
Deploy diagnostics API and frontend to Railway:
```bash
railway init
railway up
```

Configuration in `railway.json`.

### Terraform (AWS)
Infrastructure as code in `infra/terraform/`:
```bash
cd infra/terraform
terraform init
terraform apply
```

## Learning Resources

- [Smallstep: step-ca Certificate Authority Overview](https://smallstep.com/docs/step-ca/)
- [Smallstep: Certificate Authority Core Concepts](https://smallstep.com/docs/step-ca/certificate-authority-core-concepts/)
- [Smallstep: Mutual TLS Implementation Guide](https://smallstep.com/docs/mtls/)
- [Smallstep: step-ca GitHub repo](https://github.com/smallstep/certificates)

## Project Structure

```
stepdeploy-lab/
├── services/
│   ├── idp-mock/
│   ├── ca-service/
│   ├── enrollment-agent/
│   ├── protected-service/
│   └── diagnostics-api/
├── shared/
│   ├── auth/
│   ├── certs/
│   ├── logging/
│   └── tracing/
├── frontend/
├── scripts/
├── docs/
├── infra/
├── docker-compose.yml
├── .github/workflows/
└── README.md