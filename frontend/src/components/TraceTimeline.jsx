import React, { useEffect, useState } from 'react'
import { getTraces, getTraceDetail } from '../api/client'

export default function TraceTimeline() {
  const [traces, setTraces] = useState([])
  const [selectedTrace, setSelectedTrace] = useState(null)
  const [traceDetail, setTraceDetail] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchTraces = async () => {
      try {
        const response = await getTraces()
        setTraces(response.data.traces || [])
        setLoading(false)
      } catch (err) {
        console.error('Error fetching traces:', err)
        setLoading(false)
      }
    }

    fetchTraces()
    const interval = setInterval(fetchTraces, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleTraceSelect = async (trace) => {
    setSelectedTrace(trace)
    try {
      const response = await getTraceDetail(trace.trace_id)
      setTraceDetail(response.data)
    } catch (err) {
      console.error('Error fetching trace detail:', err)
    }
  }

  const getStatusColor = (status) => {
    switch(status) {
      case 'success': return '#4CAF50'
      case 'failed': return '#F44336'
      case 'in-progress': return '#2196F3'
      default: return '#999'
    }
  }

  if (loading) return <div className="card">Loading traces...</div>

  return (
    <div className="card">
      <h2>Trace Timeline</h2>
      {traces.length === 0 ? (
        <div className="empty-state">No traces recorded yet. Run enrollment to generate traces.</div>
      ) : (
        <div className="traces-container">
          <div className="traces-list">
            {traces.map((trace) => (
              <div 
                key={trace.trace_id} 
                className={`trace-item ${selectedTrace?.trace_id === trace.trace_id ? 'selected' : ''}`}
                onClick={() => handleTraceSelect(trace)}
                style={{borderLeftColor: getStatusColor(trace.status)}}
              >
                <div className="trace-header">
                  <span className="trace-id">{trace.trace_id.substring(0, 8)}</span>
                  <span className="trace-status">{trace.status}</span>
                </div>
                {trace.root_cause && <div className="trace-root-cause">Root cause: {trace.root_cause}</div>}
                <div className="trace-event-count">{trace.events?.length || 0} events</div>
              </div>
            ))}
          </div>
          
          {traceDetail && (
            <div className="trace-detail">
              <h3>Trace Details: {traceDetail.trace_id}</h3>
              <div className="trace-status-badge" style={{backgroundColor: getStatusColor(traceDetail.status)}}>
                {traceDetail.status}
              </div>
              {traceDetail.root_cause && (
                <div className="root-cause-box">
                  <strong>Root Cause:</strong> {traceDetail.root_cause}
                </div>
              )}
              <div className="events-timeline">
                {traceDetail.events?.map((event, idx) => (
                  <div key={idx} className="event-item">
                    <div className="event-time">{new Date(event.timestamp).toLocaleTimeString()}</div>
                    <div className="event-service">{event.service}</div>
                    <div className="event-type">{event.event_type}</div>
                    <div className="event-description">{event.description}</div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
