import React, { useEffect, useState } from 'react'
import { getDependencies } from '../api/client'

export default function DependencyMap() {
  const [dependencies, setDependencies] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const fetchDependencies = async () => {
      try {
        const response = await getDependencies()
        setDependencies(response.data)
        setLoading(false)
      } catch (err) {
        setError(err.message)
        setLoading(false)
      }
    }

    fetchDependencies()
    const interval = setInterval(fetchDependencies, 5000)
    return () => clearInterval(interval)
  }, [])

  if (loading) return <div className="card">Loading dependencies...</div>
  if (error) return <div className="card error">Error: {error}</div>

  const getStatusColor = (status) => {
    switch(status) {
      case 'up': return '#4CAF50'
      case 'degraded': return '#FF9800'
      case 'down': return '#F44336'
      default: return '#999'
    }
  }

  const getStatusEmoji = (status) => {
    switch(status) {
      case 'up': return '✓'
      case 'degraded': return '⚠'
      case 'down': return '✗'
      default: return '?'
    }
  }

  return (
    <div className="card">
      <h2>Service Dependencies</h2>
      <div className="status-badge" style={{backgroundColor: getStatusColor(dependencies.status)}}>
        Overall: {dependencies.status.toUpperCase()}
      </div>
      <div className="services-grid">
        {dependencies.services && dependencies.services.map((service, idx) => (
          <div key={idx} className="service-card" style={{borderLeftColor: getStatusColor(service.status)}}>
            <div className="service-header">
              <span className="status-emoji">{getStatusEmoji(service.status)}</span>
              <span className="service-name">{service.name}</span>
              <span className="status-label">{service.status}</span>
            </div>
            <div className="service-url">{service.url}</div>
            <div className="service-message">{service.message}</div>
            <div className="service-timestamp">Last check: {new Date(service.last_check).toLocaleTimeString()}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
