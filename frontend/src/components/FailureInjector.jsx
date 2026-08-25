import React, { useState } from 'react'
import { testEnroll } from '../api/client'

export default function FailureInjector() {
  const [enrollmentResult, setEnrollmentResult] = useState(null)
  const [loading, setLoading] = useState(false)

  const handleTestEnroll = async () => {
    setLoading(true)
    try {
      const response = await testEnroll()
      setEnrollmentResult(response.data)
    } catch (err) {
      setEnrollmentResult({
        status: 'error',
        message: err.message
      })
    } finally {
      setLoading(false)
    }
  }

  const getStatusColor = (status) => {
    switch(status) {
      case 'success': return '#4CAF50'
      case 'error': return '#F44336'
      case 'failed': return '#F44336'
      default: return '#2196F3'
    }
  }

  return (
    <div className="card">
      <h2>Test Enrollment Flow</h2>
      <div className="enrollment-section">
        <button 
          className="btn-primary"
          onClick={handleTestEnroll}
          disabled={loading}
        >
          {loading ? 'Running Enrollment...' : 'Run Enrollment Test'}
        </button>

        {enrollmentResult && (
          <div className="enrollment-result" style={{borderLeftColor: getStatusColor(enrollmentResult.status)}}>
            <h3>Enrollment Result</h3>
            <div className="result-grid">
              <div className="result-item">
                <label>Status</label>
                <div style={{color: getStatusColor(enrollmentResult.status)}}>
                  {enrollmentResult.status ? enrollmentResult.status.toUpperCase() : 'UNKNOWN'}
                </div>
              </div>
              <div className="result-item">
                <label>Trace ID</label>
                <div className="trace-id">{enrollmentResult.trace_id?.substring(0, 16)}...</div>
              </div>
              <div className="result-item">
                <label>Token</label>
                <div style={{color: enrollmentResult.token_ok ? '#4CAF50' : '#F44336'}}>
                  {enrollmentResult.token_ok ? '✓ OK' : '✗ Failed'}
                </div>
              </div>
              <div className="result-item">
                <label>Certificate</label>
                <div style={{color: enrollmentResult.cert_ok ? '#4CAF50' : '#F44336'}}>
                  {enrollmentResult.cert_ok ? '✓ OK' : '✗ Failed'}
                </div>
              </div>
              <div className="result-item">
                <label>Access</label>
                <div style={{color: enrollmentResult.access_ok ? '#4CAF50' : '#F44336'}}>
                  {enrollmentResult.access_ok ? '✓ OK' : '✗ Failed'}
                </div>
              </div>
            </div>
            {enrollmentResult.message && (
              <div className="result-message">{enrollmentResult.message}</div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
