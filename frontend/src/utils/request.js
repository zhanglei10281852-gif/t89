import axios from 'axios'
import { message } from 'ant-design-vue'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    if (error.response) {
      if (error.response.status === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        window.location.href = '/login'
        message.error('登录已过期，请重新登录')
      } else if (error.response.data && error.response.data.error) {
        message.error(error.response.data.error)
      } else {
        message.error('请求失败: ' + error.message)
      }
    }
    return Promise.reject(error)
  }
)

export default request
