import React, { useEffect, useState } from 'react'
import { getIncidents, getIncidentDetail, resolveIncident } from '../api/client'

export default function IncidentTracker() {
  const [incidents, setIncidents] = useState([])
  const [selectedIncident, setSelectedIncident] = useState(null)
  const [incidentDetail, setIncidentDetail] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchIncidents = async () => {
      try {
        const response = await getIncidents()
        setIncidents(response.data.incidents || [])
        setLoading(false)
      } catch (err) {
        console.error('Error fetching incidents:', err)
        setLoading(false)
      }
    }

    fetchIncidents()
    const interval = setInterval(fetchIncidents, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleIncidentSelect = async (incident) => {
    setSelectedIncident(incident)
    try {
      const response = await getIncidentDetail(incident.id)
      setIncidentDetail(response.data)
    } catch (err) {
      console.error('Error fetching incident detail:', err)
    }
  }

  const handleResolveIncident = async (incidentId) => {
    try {
      await resolveIncident(incidentId)
      // Refresh incidents
      const response = await getIncidents()
      setIncidents(response.data.incidents || [])
      setIncidentDetail(null)
      setSelectedIncident(null)
    } catch (err) {
      console.error('Error resolving incident:', err)
    }
  }

  const getSeverityColor = (severity) => {
    switch(severity) {
      case 'critical': return '#F44336'
      case 'high': return '#FF6F00'
      case 'medium': return '#FBC02D'
      case 'low': return '#4CAF50'
      default: return '#999'
    }
  }

  const getSLAStatus = (incident) => {
    if (incident.sla_breached) {
      return <span style={{color: '#F44336'}}>⚠ SLA BREACHED</span>
    }
    return <span style={{color: '#4CAF50'}}>✓ Within SLA</span>
  }

  if (loading) return <div className="card">Loading incidents...</div>

  return (
    <div className="card">
      <h2>Incident Tracker</h2>
      {incidents.length === 0 ? (
        <div className="empty-state">No incidents. System is healthy.</div>
      ) : (
        <div className="incidents-container">
          <div className="incidents-list">
            {incidents.map((incident) => (
              <div 
                key={incident.id} 
                className={`incident-item ${selectedIncident?.id === incident.id ? 'selected' : ''}`}
                onClick={() => handleIncidentSelect(incident)}
                style={{borderLeftColor: getSeverityColor(incident.severity)}}
              >
                <div className="incident-header">
                  <span className="incident-id">{incident.id}</span>
                  <span className="incident-severity" style={{backgroundColor: getSeverityColor(incident.severity)}}>
                    {incident.severity.toUpperCase()}
                  </span>
                </div>
                <div className="incident-title">{incident.title}</div>
                <div className="incident-timestamp">
                  Created: {new Date(incident.created_at).toLocaleString()}
                  {incident.resolved_at && ` • Resolved: ${new Date(incident.resolved_at).toLocaleString()}`}
                </div>
              </div>
            ))}
          </div>
          
          {incidentDetail && (
            <div className="incident-detail">
              <h3>{incidentDetail.title}</h3>
              <div className="incident-meta">
                <div><strong>ID:</strong> {incidentDetail.id}</div>
                <div><strong>Trace ID:</strong> {incidentDetail.trace_id}</div>
                <div><strong>Severity:</strong> <span style={{color: getSeverityColor(incidentDetail.severity)}}>{incidentDetail.severity.toUpperCase()}</span></div>
                <div><strong>SLA Target:</strong> {incidentDetail.sla_target_minutes} minutes</div>
                <div><strong>Status:</strong> {getSLAStatus(incidentDetail)}</div>
              </div>
              
              <div className="incident-description">
                <strong>Description:</strong>
                <p>{incidentDetail.description}</p>
              </div>

              {incidentDetail.root_cause && (
                <div className="root-cause-box">
                  <strong>Root Cause:</strong>
                  <p>{incidentDetail.root_cause}</p>
                </div>
              )}

              {incidentDetail.runbook && (
                <div className="runbook-box">
                  <strong>Runbook:</strong>
                  <p>{incidentDetail.runbook}</p>
                </div>
              )}

              {!incidentDetail.resolved_at && (
                <button 
                  className="btn-resolve"
                  onClick={() => handleResolveIncident(incidentDetail.id)}
                >
                  Resolve Incident
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
