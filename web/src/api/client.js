export async function request(url, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  }

  const res = await fetch(url, {
    ...options,
    headers,
  })

  let data = null
  const contentType = res.headers.get('content-type')
  if (contentType && contentType.includes('application/json')) {
    data = await res.json()
  } else {
    data = await res.text()
  }

  if (!res.ok) {
    const errorMsg = (data && data.error) ? data.error : (typeof data === 'string' ? data : res.statusText)
    throw new Error(errorMsg || `Request failed with status ${res.status}`)
  }

  return data
}
