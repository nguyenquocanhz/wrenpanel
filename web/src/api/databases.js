import { request } from './client'

export const databasesApi = {
  list(accountId) {
    const q = accountId ? `?account_id=${accountId}` : ''
    return request(`/api/databases${q}`)
  },
  create(payload) {
    return request('/api/databases', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  delete(id) {
    return request(`/api/databases/${id}`, {
      method: 'DELETE',
    })
  },
}
