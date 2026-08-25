import React, { useState } from 'react'
import Home from './pages/Home'
import IncidentsPage from './pages/IncidentsPage'
import './styles/index.css'

export default function App() {
  const [activePage, setActivePage] = useState('home')

  return (
    <div className="app">
      <header className="header">
        <h1>StepDeploy Lab Dashboard</h1>
        <p>Multi-service certificate deployment with failure injection and diagnostics</p>
      </header>

      <nav className="nav">
        <button
          className={`nav-btn ${activePage === 'home' ? 'active' : ''}`}
          onClick={() => setActivePage('home')}
        >
          Dashboard
        </button>
        <button
          className={`nav-btn ${activePage === 'incidents' ? 'active' : ''}`}
          onClick={() => setActivePage('incidents')}
        >
          Incidents
        </button>
      </nav>

      {activePage === 'home' && <Home />}
      {activePage === 'incidents' && <IncidentsPage />}

      <footer className="footer">
        <p>Powered by Smallstep | stonesmith.dev portfolio</p>
      </footer>
    </div>
  )
}
