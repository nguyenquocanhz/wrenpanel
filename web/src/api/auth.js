import { request } from './client'

export const authApi = {
  login(username, password) {
    return request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  },
  getMe() {
    return request('/api/auth/me')
  },
}
