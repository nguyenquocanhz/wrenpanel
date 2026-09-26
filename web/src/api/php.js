import { request } from './client'

export const phpApi = {
  listVersions() {
    return request('/api/php/versions')
  },
  createVersion(payload) {
    return request('/api/php/versions', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  deleteVersion(id) {
    return request(`/api/php/versions/${id}`, {
      method: 'DELETE',
    })
  },
  reload(version) {
    return request('/api/php/reload', {
      method: 'POST',
      body: JSON.stringify({ version }),
    })
  },
}
