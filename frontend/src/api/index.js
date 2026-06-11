import request from '../utils/request'

export const api = {
  auth: {
    login: (data) => request.post('/auth/login', data),
    me: () => request.get('/auth/me')
  },
  users: {
    list: () => request.get('/users'),
    planners: () => request.get('/users/planners')
  },
  serviceItems: {
    list: () => request.get('/service-items'),
    core: () => request.get('/service-items/core')
  },
  packages: {
    list: () => request.get('/packages'),
    get: (id) => request.get(`/packages/${id}`),
    create: (data) => request.post('/packages', data),
    update: (id, data) => request.put(`/packages/${id}`, data),
    delete: (id) => request.delete(`/packages/${id}`)
  },
  customers: {
    list: (params) => request.get('/customers', { params }),
    get: (id) => request.get(`/customers/${id}`),
    create: (data) => request.post('/customers', data),
    update: (id, data) => request.put(`/customers/${id}`, data),
    updateStatus: (id, status) => request.patch(`/customers/${id}/status`, { status }),
    delete: (id) => request.delete(`/customers/${id}`),
    funnel: () => request.get('/customers/funnel')
  },
  quotes: {
    list: (params) => request.get('/quotes', { params }),
    get: (id) => request.get(`/quotes/${id}`),
    create: (data) => request.post('/quotes', data),
    createVersion: (id) => request.post(`/quotes/${id}/version`),
    update: (id, data) => request.put(`/quotes/${id}`, data),
    confirm: (id) => request.post(`/quotes/${id}/confirm`),
    delete: (id) => request.delete(`/quotes/${id}`)
  },
  schedules: {
    availability: (date) => request.get('/schedules/availability', { params: { date } }),
    checkConflict: (data) => request.post('/schedules/check-conflict', data),
    calendar: (year) => request.get('/schedules/calendar', { params: { year } }),
    luckyDays: () => request.get('/schedules/lucky-days'),
    mySchedule: (month) => request.get('/schedules/mine', { params: { month } })
  },
  contracts: {
    list: () => request.get('/contracts'),
    get: (id) => request.get(`/contracts/${id}`),
    create: (data) => request.post('/contracts', data),
    update: (id, data) => request.put(`/contracts/${id}`, data),
    delete: (id) => request.delete(`/contracts/${id}`)
  },
  stats: {
    monthly: (year) => request.get('/stats/monthly', { params: { year } }),
    conversion: () => request.get('/stats/conversion'),
    plannerLoad: (year) => request.get('/stats/planner-load', { params: { year } }),
    packageRank: () => request.get('/stats/package-rank'),
    serviceRank: () => request.get('/stats/service-rank'),
    luckyComparison: (year) => request.get('/stats/lucky-comparison', { params: { year } })
  }
}
