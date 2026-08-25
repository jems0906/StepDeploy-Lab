import React, { useState } from 'react'
import DependencyMap from '../components/DependencyMap'
import TraceTimeline from '../components/TraceTimeline'
import FailureInjector from '../components/FailureInjector'
import RunbookPanel from '../components/RunbookPanel'

export default function Home() {
  const [activeTab, setActiveTab] = useState('overview')

  return (
    <>
      <nav className="nav">
        <button
          className={`nav-btn ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          Overview
        </button>
        <button
          className={`nav-btn ${activeTab === 'traces' ? 'active' : ''}`}
          onClick={() => setActiveTab('traces')}
        >
          Traces
        </button>
        <button
          className={`nav-btn ${activeTab === 'test' ? 'active' : ''}`}
          onClick={() => setActiveTab('test')}
        >
          Test
        </button>
        <button
          className={`nav-btn ${activeTab === 'runbooks' ? 'active' : ''}`}
          onClick={() => setActiveTab('runbooks')}
        >
          Runbooks
        </button>
      </nav>

      <main className="main">
        {activeTab === 'overview' && (
          <div className="tab-content">
            <DependencyMap />
          </div>
        )}

        {activeTab === 'traces' && (
          <div className="tab-content">
            <TraceTimeline />
          </div>
        )}

        {activeTab === 'test' && (
          <div className="tab-content">
            <FailureInjector />
          </div>
        )}

        {activeTab === 'runbooks' && (
          <div className="tab-content">
            <RunbookPanel />
          </div>
        )}
      </main>
    </>
  )
}
