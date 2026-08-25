import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api'

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

export const getDependencies = () => client.get('/dependencies')
export const getTraces = () => client.get('/traces')
export const getTraceDetail = (traceId) => client.get(`/traces/detail?trace_id=${traceId}`)
export const getIncidents = () => client.get('/incidents')
export const getIncidentDetail = (incidentId) => client.get(`/incidents/detail?incident_id=${incidentId}`)
export const createIncident = (data) => client.post('/incidents', data)
export const resolveIncident = (incidentId) => client.post(`/incidents/resolve?incident_id=${incidentId}`)
export const testEnroll = () => client.post('/enroll')

export default client
