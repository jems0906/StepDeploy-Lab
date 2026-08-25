# Escalation Playbook

## When to Escalate

### Tier 1 → Tier 2 Criteria
Escalate to engineering if:
1. Issue persists after running all runbook steps
2. Multiple services affected simultaneously  
3. Data corruption suspected
4. Security or authentication bypass
5. Performance degradation >50% from baseline

### Incident Severity

**Critical (Escalate Immediately)**
- All services down (enrollment impossible)
- Certificate issuance completely broken
- mTLS verification fails for all clients
- Data loss or corruption detected

**High (Escalate if unresolved in 15 min)**
- One core service down
- Token generation broken
- CA unavailable but IdP working
- Partial functionality loss

**Medium (Escalate if unresolved in 60 min)**
- One optional service degraded
- Performance issues (>20% latency increase)
- Intermittent failures
- Non-critical feature broken

**Low (Escalate if unresolved in 24 hr)**
- Cosmetic issues
- Documentation outdated
- Rare edge case failures
- Development-only features broken

## Escalation Template

When escalating, provide:

```
INCIDENT ESCALATION

Severity: [CRITICAL|HIGH|MEDIUM|LOW]
Time to Escalate: [HH:MM]
Runbooks Attempted: [list]

Affected Service: [service-name]
Failed Step: [step description]
Error Message: [exact error or log snippet]
Frequency: [always|intermittent|single occurrence]

Environment Details:
- Docker version: $(docker --version)
- Go version: $(go version)
- Node version: $(node --version)
- System memory: $(free -h)
- Disk space: $(df -h)

Recent Actions:
- [action and timestamp]
- [action and timestamp]

Attached: 
- logs-YYYYMMDD-HHMMSS/ (from ./scripts/collect-logs.sh)
- /diagnostics/traces (from dashboard)
- /diagnostics/incidents (from dashboard)

Questions for Engineering:
- [specific technical question]
- [config parameter needed]
```

## Communication Channels

### Priority by Severity

| Severity | Channel | Response Time |
|----------|---------|---|
| Critical | Page (PagerDuty) + Slack #incidents | 15 min |
| High | Slack #incidents + email | 30 min |
| Medium | Email + issue tracker | 2 hours |
| Low | Issue tracker | 24 hours |

### Escalation Contacts

**On-Call Engineer**: 
- Check rotation calendar: `/on-call`
- Page: `incident escalate <on-call-engineer>`
- Slack: Mention in #on-call

**Engineering Manager**:
- Email: engineering-manager@stepdeploy.internal
- Slack: @eng-manager

**Product Team**:
- Slack: #product-ops
- For customer-impacting issues

## Information to Collect

### Before Escalating
1. Run health checks: `./scripts/run-healthcheck.sh`
2. Collect logs: `./scripts/collect-logs.sh`
3. Get current traces: Dashboard → Traces → Export
4. Get incidents: Dashboard → Incidents → Export
5. Document exact reproduction steps
6. Identify when issue started
7. Check recent changes (git history)

### During Call/Handoff
1. Explain what you've already tried
2. Provide reproduction steps
3. Share logs and traces
4. Describe expected vs actual behavior
5. State any urgency/customer impact

## Post-Incident Process

After resolving escalated incident:
1. Document root cause in issue tracker
2. Update relevant runbook
3. Add test case to prevent regression
4. Schedule follow-up engineering review
5. Brief team in retrospective
6. Update monitoring/alerting rules

## Examples

### Example 1: Critical - All Services Down
```
INCIDENT: All services down, enrollment impossible

Environment: Production (AWS)
Severity: CRITICAL
Time to Detect: 2 minutes
Time to Escalate: 5 minutes

Error:
docker-compose: bind: address already in use
Port conflict on :8001

Attempted:
✓ Checked netstat - no conflicts found
✓ Restarted Docker daemon
✓ Checked disk space (45% free)
✓ Attempted full rebuild - same error persists

Status: Cannot recover. Escalating to on-call.
```

### Example 2: Medium - CA Certificate Signing Slow  
```
INCIDENT: Certificate signing taking 30+ seconds

Severity: MEDIUM  
Frequency: Intermittent (50% of requests)
Detection: Traces show cert-request event taking 25-30 seconds

Environment:
- 2GB RAM container, 1 CPU
- Load: 50 concurrent requests
- Docker stats show CPU at 95%

Attempted:
✓ Reviewed CA logs - no errors
✓ Checked CA service metrics
✓ Verified certificate generation works in isolation
✓ Increased container memory to 4GB - still slow

Next Steps: Needs CPU profiling to identify bottleneck.
Escalating for profiling support.
```

## Runbook Links

- [expired-oauth-token](runbooks/expired-oauth-token.md) - Token validation failures
- [bad-client-cert](runbooks/bad-client-cert.md) - Certificate issues  
- [missing-ca-trust-chain](runbooks/missing-ca-trust-chain.md) - Trust store problems
- [service-unavailable](runbooks/service-unavailable.md) - Service crashes
- [DNS mismatch and network issues](runbooks/service-unavailable.md)
- [clock-skew and timing issues](runbooks/expired-oauth-token.md)

---

Last Updated: 2024
Maintained By: SRE Team
Review Cycle: Quarterly
