const API_BASE_URL = 'http://localhost:8080/api'

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.error || response.statusText || 'Request failed')
  }

  return response.json()
}

export const adminApi = {
  login: credentials => request('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials)
  }),
  dashboard: () => request('/dashboard'),
  users: () => request('/users'),
  createUser: user => request('/users', {
    method: 'POST',
    body: JSON.stringify(user)
  }),
  updateUser: (id, user) => request(`/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(user)
  }),
  resetPassword: (id, password) => request(`/users/${id}/password`, {
    method: 'PATCH',
    body: JSON.stringify({ password })
  }),
  deleteUser: id => request(`/users/${id}`, { method: 'DELETE' }),
  roles: () => request('/roles'),
  createRole: role => request('/roles', {
    method: 'POST',
    body: JSON.stringify(role)
  }),
  updateRole: (id, role) => request(`/roles/${id}`, {
    method: 'PUT',
    body: JSON.stringify(role)
  }),
  deleteRole: id => request(`/roles/${id}`, { method: 'DELETE' }),
  permissions: () => request('/permissions'),
  modules: () => request('/modules'),
  updateModule: (id, module) => request(`/modules/${id}`, {
    method: 'PUT',
    body: JSON.stringify(module)
  }),
  settings: () => request('/settings'),
  saveSettings: settings => request('/settings', {
    method: 'PUT',
    body: JSON.stringify(settings)
  }),
  auditLogs: () => request('/audit-logs')
}
