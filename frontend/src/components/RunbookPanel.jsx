import React from 'react'

export default function RunbookPanel() {
  const runbooks = [
    {
      id: 'expired-token',
      title: 'Expired OAuth Token',
      description: 'OAuth token has expired and needs to be refreshed',
      steps: [
        '1. Check IdP service is running and healthy',
        '2. Verify token TTL configuration (expires_in)',
        '3. Request new token from IdP',
        '4. Retry enrollment with new token'
      ]
    },
    {
      id: 'audience-mismatch',
      title: 'Token Audience Mismatch',
      description: 'OAuth token audience does not match CA service audience',
      steps: [
        '1. Verify CA service audience configuration',
        '2. Check IdP token issuance configuration',
        '3. Ensure audience claim in token matches CA_SERVICE audience',
        '4. Regenerate token with correct audience'
      ]
    },
    {
      id: 'bad-cert',
      title: 'Certificate Validation Failed',
      description: 'CA rejected certificate signing request',
      steps: [
        '1. Verify CA service is running',
        '2. Check certificate CSR format',
        '3. Verify CA trust chain is complete',
        '4. Review CA logs for detailed error'
      ]
    },
    {
      id: 'service-unavailable',
      title: 'Protected Service Unavailable',
      description: 'Cannot connect to protected service',
      steps: [
        '1. Check if protected-service container is running',
        '2. Verify mTLS configuration',
        '3. Check firewall rules and network connectivity',
        '4. Review service logs for startup errors'
      ]
    }
  ]

  return (
    <div className="card">
      <h2>Runbook Reference</h2>
      <div className="runbooks-grid">
        {runbooks.map((runbook) => (
          <div key={runbook.id} className="runbook-card">
            <h3>{runbook.title}</h3>
            <p className="runbook-desc">{runbook.description}</p>
            <div className="runbook-steps">
              {runbook.steps.map((step, idx) => (
                <div key={idx} className="step">{step}</div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
