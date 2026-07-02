import { request } from './client'
import type {
  AppSetting,
  AuthResponse,
  Category,
  CategoryType,
  Dashboard,
  Expense,
  SavingsTargetType,
  User,
  UserStatus,
} from '../types'

export const authApi = {
  register: (email: string, password: string) =>
    request<AuthResponse>('/api/auth/register', { method: 'POST', body: { email, password } }),
  login: (email: string, password: string) =>
    request<AuthResponse>('/api/auth/login', { method: 'POST', body: { email, password } }),
  logout: () => request<{ status: string }>('/api/auth/logout', { method: 'POST' }),
}

export const meApi = {
  get: () => request<User>('/api/me'),
  setIncome: (monthlyIncome: string) =>
    request<User>('/api/me', { method: 'PATCH', body: { monthlyIncome } }),
}

export const categoryApi = {
  list: () => request<Category[]>('/api/categories'),
  create: (name: string, type: CategoryType) =>
    request<Category>('/api/categories', { method: 'POST', body: { name, type } }),
  update: (id: string, name: string, type: CategoryType) =>
    request<Category>(`/api/categories/${id}`, { method: 'PATCH', body: { name, type } }),
  remove: (id: string) => request<{ status: string }>(`/api/categories/${id}`, { method: 'DELETE' }),
}

export interface ExpenseInput {
  amount: string
  note: string | null
  spentOn: string
  categoryId: string
}

export const expenseApi = {
  list: (month: string) => request<Expense[]>(`/api/expenses?month=${month}`),
  create: (input: ExpenseInput) => request<Expense>('/api/expenses', { method: 'POST', body: input }),
  update: (id: string, input: ExpenseInput) =>
    request<Expense>(`/api/expenses/${id}`, { method: 'PATCH', body: input }),
  remove: (id: string) => request<{ status: string }>(`/api/expenses/${id}`, { method: 'DELETE' }),
}

export const dashboardApi = {
  get: (month: string) => request<Dashboard>(`/api/dashboard?month=${month}`),
}

export const adminApi = {
  listUsers: () => request<User[]>('/api/admin/users'),
  setUserStatus: (id: string, status: UserStatus) =>
    request<User>(`/api/admin/users/${id}/status`, { method: 'PATCH', body: { status } }),
  deleteUser: (id: string) =>
    request<{ status: string }>(`/api/admin/users/${id}`, { method: 'DELETE' }),
  getSettings: () => request<AppSetting>('/api/admin/settings'),
  updateSettings: (savingsTargetType: SavingsTargetType, savingsTargetValue: string) =>
    request<AppSetting>('/api/admin/settings', {
      method: 'PUT',
      body: { savingsTargetType, savingsTargetValue },
    }),
}
