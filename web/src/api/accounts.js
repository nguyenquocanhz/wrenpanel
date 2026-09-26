import { request } from './client'

export const accountsApi = {
  list() {
    return request('/api/accounts')
  },
  create(payload) {
    return request('/api/accounts', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  delete(id) {
    return request(`/api/accounts/${id}`, {
      method: 'DELETE',
    })
  },
}
