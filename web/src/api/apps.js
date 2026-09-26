import { request } from './client'

export const appsApi = {
  list(accountId, kind) {
    const params = new URLSearchParams()
    if (accountId) params.append('account_id', accountId)
    if (kind) params.append('kind', kind)
    const q = params.toString() ? `?${params.toString()}` : ''
    return request(`/api/apps${q}`)
  },
  get(id) {
    return request(`/api/apps/${id}`)
  },
  create(payload) {
    return request('/api/apps', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  action(id, actionName) {
    return request(`/api/apps/${id}/action`, {
      method: 'POST',
      body: JSON.stringify({ action: actionName }),
    })
  },
  delete(id) {
    return request(`/api/apps/${id}`, {
      method: 'DELETE',
    })
  },
  // Node Runtimes
  listNodeVersions() {
    return request('/api/node/versions')
  },
  createNodeVersion(payload) {
    return request('/api/node/versions', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  deleteNodeVersion(id) {
    return request(`/api/node/versions/${id}`, {
      method: 'DELETE',
    })
  },
  // Python Runtimes
  listPythonVersions() {
    return request('/api/python/versions')
  },
  createPythonVersion(payload) {
    return request('/api/python/versions', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  deletePythonVersion(id) {
    return request(`/api/python/versions/${id}`, {
      method: 'DELETE',
    })
  },
}
