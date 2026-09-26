import { request } from './client'

export const ftpApi = {
  list(accountId) {
    const q = accountId ? `?account_id=${accountId}` : ''
    return request(`/api/ftp${q}`)
  },
  create(payload) {
    return request('/api/ftp', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  delete(id) {
    return request(`/api/ftp/${id}`, {
      method: 'DELETE',
    })
  },
  getPmaStatus() {
    return request('/api/db-manager/status')
  },
  launchPma() {
    return request('/api/db-manager/launch', {
      method: 'POST',
    })
  },
}
