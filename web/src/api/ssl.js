import { request } from './client'

export const sslApi = {
  listCerts() {
    return request('/api/ssl/certs')
  },
  issueCert(vhostId, email) {
    return request('/api/ssl/issue', {
      method: 'POST',
      body: JSON.stringify({ vhost_id: vhostId, email }),
    })
  },
}
