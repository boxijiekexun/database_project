import axios from 'axios'

const request = axios.create({
  baseURL: 'http://localhost:8080',
  timeout: 15000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const payload = response.data
    if (payload && Object.prototype.hasOwnProperty.call(payload, 'code') && payload.code !== 0) {
      const message = payload.message || '请求失败'
      window.alert(message)
      return Promise.reject(new Error(message))
    }
    return payload && Object.prototype.hasOwnProperty.call(payload, 'data') ? payload.data : payload
  },
  (error) => {
    const message = error?.response?.data?.message || error.message || '网络请求失败'
    window.alert(message)
    return Promise.reject(error)
  }
)

export function clearAuth() {
  localStorage.removeItem('token')
  localStorage.removeItem('role')
}

export default request
