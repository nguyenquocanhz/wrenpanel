import { request } from './client'

export const systemApi = {
  getStatus() {
    return request('/api/system/status')
  },
  listAuditLogs(limit = 100) {
    return request(`/api/audit-logs?limit=${limit}`)
  },
}
