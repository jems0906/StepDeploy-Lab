# Architecture

## Overview

StepDeploy Lab is a multi-service Go system that demonstrates a Smallstep-style customer deployment with:

- OAuth/OIDC identity provider (IdP)
- Certificate authority (CA) for issuing short-lived X.509 client certificates
- Enrollment agent (device-side enrollment workflow)
- Protected mTLS-enabled service
- Diagnostics API with health checks, trace collection, and incident tracking
- React dashboard for visualizing system state and debugging

## Service Architecture

```
┌─────────────┐         ┌──────────┐          ┌────────────┐
│  Device    │────────▶ │   IdP    │          │ Diagnostics│
│ Enrollment │ GET /   │  Mock    │          │   API      │
│   Agent    │ token   │          │          │            │
│            │         └──────────┘          └────────────┘
│            │              │                      ▲
│            │              │                      │
│            │         GET /token            Monitors all
│            │              │                      │
│            │         ┌─────────────┐         Endpoints
│            │         │     CA      │             │
│            │────────▶│  Service    │─────────────┘
│            │ POST    │             │
│            │ /certs/sign          │
│            │         └─────────────┘
│            │
│            │    mTLS with
│            │    Client Cert
│            │
│            ▼
│      ┌──────────────┐
│      │  Protected   │
│      │  Service     │
│      │  (mTLS)      │
│      └──────────────┘
```

## Services

### IdP Mock (Port 8001)
- **Role**: OAuth/OIDC token issuance
- **Technology**: Go, gorilla/mux
- **Endpoints**:
  - `POST /token` - Issue OAuth token
  - `GET /userinfo` - Get user information
  - `GET /health` - Health check
  - `POST /test/inject-failure` - Inject test failures
  - `POST /test/reset-failures` - Reset failures

**Key Features**:
- Simulates Okta/Google Workspace OAuth behavior
- Configurable token TTL
- Supports audience-specific tokens
- Can inject expired-token and wrong-audience failures

### CA Service (Port 8002)
- **Role**: X.509 certificate issuance
- **Technology**: Go, crypto/x509
- **Endpoints**:
  - `POST /certs/sign` - Sign certificate request with OAuth token
  - `GET /certs/root` - Get root CA certificate
  - `GET /health` - Health check
  - `POST /test/inject-failure` - Inject test failures
  - `POST /test/reset-failures` - Reset failures

**Key Features**:
- Manages root CA and intermediate CA
- OAuth-based provisioner validates tokens before issuance
- Issues short-lived client certificates (1-7 days)
- Can inject bad-client-cert and missing-ca-trust failures
- Returns full CA chain (intermediate + root) to client

### Enrollment Agent (Port 8003)
- **Role**: Device-side enrollment workflow
- **Technology**: Go, HTTP client
- **Endpoints**:
  - `POST /enroll` - Execute full enrollment flow
  - `GET /health` - Health check
  - `POST /test/inject-failure` - Inject test failures
  - `POST /test/reset-failures` - Reset failures

**Key Features**:
- Implements 3-step enrollment flow:
  1. Get OAuth token from IdP
  2. Request client certificate from CA
  3. Access protected service with certificate
- Creates distributed trace for entire workflow
- Can inject failures at each step
- Returns enrollment status and trace ID

**Enrollment Workflow**:
```
Start ──▶ Get Token ──▶ Get Cert ──▶ Access Resource ──▶ Success
                         │               │
                         ▼               ▼
                    If expired      If denied
                         │               │
                         ▼               ▼
                      Failed ◀────────────┘
```

### Protected Service (Port 8004)
- **Role**: mTLS-protected resource server
- **Technology**: Go, crypto/tls
- **Endpoints**:
  - `GET /api/resource` - Protected resource (requires client cert)
  - `GET /health` - Health check
  - `POST /test/inject-failure` - Inject test failures
  - `POST /test/reset-failures` - Reset failures

**Key Features**:
- Requires valid client certificate via mTLS
- Extracts client certificate DN from connection
- Verifies certificate chain
- Can simulate service unavailability

### Diagnostics API (Port 8080)
- **Role**: Central diagnostics and monitoring hub
- **Technology**: Go, gorilla/mux
- **Endpoints**:
  - `GET /health` - Service health
  - `GET /dependencies` - Dependency health map
  - `GET /traces` - List all traces
  - `GET /traces/detail?trace_id=XXX` - Get trace details
  - `GET /incidents` - List incidents
  - `GET /incidents/detail?incident_id=XXX` - Get incident details
  - `POST /incidents` - Create new incident
  - `POST /incidents/resolve?incident_id=XXX` - Resolve incident
  - `POST /enroll` - Trigger enrollment test

**Key Features**:
- Periodic health checks of all services (every 10 seconds)
- Aggregates dependency status
- Stores and retrieves distributed traces
- Tracks incidents with SLA management
- Checks SLA breaches based on incident severity
- CORS enabled for frontend access

## Frontend (Port 5173)
- **Technology**: React, Vite, Axios
- **Components**:
  - Dependency Map - Service health visualization
  - Trace Timeline - Distributed trace explorer
  - Incident Tracker - SLA-aware incident management
  - Failure Injector - Test enrollment flow
  - Runbook Panel - Troubleshooting guides

## Distributed Tracing

Each enrollment request generates a trace with:
- Unique trace ID
- Timeline of events across services
- Event metadata (timestamps, service names)
- Final status and root cause (if failed)

Events are created at each stage:
1. `enrollment-agent:start` - Enrollment initiated
2. `enrollment-agent:token-request` - Token requested
3. `enrollment-agent:token-ok` or `enrollment-agent:token-failed`
4. `enrollment-agent:cert-request` - Certificate requested
5. `enrollment-agent:cert-ok` or `enrollment-agent:cert-request-error`
6. `enrollment-agent:protected-access` - Accessing resource
7. `enrollment-agent:success` or `enrollment-agent:protected-access-error`

## Incident Tracking

Incidents are created when enrollment fails with:
- ID (auto-generated)
- Trace ID (linked to failure trace)
- Severity (critical, high, medium, low)
- SLA Target (5 min for critical, 15 min for others)
- Root Cause Analysis
- Recommended Runbook
- SLA Breach Detection

## Data Flow

```
Device ─────────────▶ IdP ──────────────▶ Diagnostics
  │                   │                      ▲
  │                   │                      │
  ▼                   ▼                      │
 CA ◀─────────────── Enrolls ─────────────── │
  │                   │                      │
  │                   ▼                      │
  └────▶ Protected   ✓ Stores Trace ────────┘
         Service     ✓ Creates Incident
                     ✓ Checks SLA
```

## Failure Points

1. **IdP failures**:
   - Service down/unavailable
   - Expired tokens
   - Wrong audience in token
   - Token validation failures

2. **CA failures**:
   - Service down/unavailable
   - Invalid/expired tokens
   - CSR validation errors
   - CA cert chain issues

3. **Protected Service failures**:
   - Service down/unavailable
   - Client cert validation failures
   - Certificate expired
   - Certificate revoked

4. **Network failures**:
   - DNS mismatch
   - Service endpoint misconfiguration
   - Network connectivity issues

5. **Timing failures**:
   - Clock skew between components
   - Certificate validity window mismatches

## Technology Stack

- **Services**: Go 1.22, net/http, gorilla/mux, crypto/x509
- **Frontend**: React 18, Vite, Axios
- **Containerization**: Docker, Docker Compose
- **Orchestration**: Railway (cloud deployment)
- **IaC**: Terraform
- **CI/CD**: GitHub Actions
