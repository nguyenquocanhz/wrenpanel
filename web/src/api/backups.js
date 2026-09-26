import { request } from './client'

export const backupsApi = {
  list(accountId) {
    const q = accountId ? `?account_id=${accountId}` : ''
    return request(`/api/backups${q}`)
  },
  create(payload) {
    return request('/api/backups', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
