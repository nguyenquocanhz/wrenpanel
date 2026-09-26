import { request } from './client'

export const vhostsApi = {
  list(accountId) {
    const query = accountId ? `?account_id=${accountId}` : ''
    return request(`/api/vhosts${query}`)
  },
  get(id) {
    return request(`/api/vhosts/${id}`)
  },
  create(payload) {
    return request('/api/vhosts', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  getDeletionImpact(id) {
    return request(`/api/vhosts/${id}/deletion-impact`)
  },
  delete(id, deleteFiles = false) {
    return request(`/api/vhosts/${id}?delete_files=${deleteFiles}`, {
      method: 'DELETE',
    })
  },
}
