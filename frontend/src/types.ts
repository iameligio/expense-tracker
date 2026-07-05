export type Role = 'member' | 'admin'
export type UserStatus = 'active' | 'suspended' | 'banned'
export type CategoryType = 'fixed' | 'variable' | 'wants' | 'debts'
export type IncomeSource = 'salary' | 'side_project' | 'other'
export type SavingsTargetType = 'percent' | 'fixed'

export interface User {
  id: string
  email: string
  role: Role
  status: UserStatus
  monthlyIncome: string
}

export interface AuthResponse {
  accessToken: string
  expiresIn: number
  user: User
}

export interface Category {
  id: string
  name: string
  type: CategoryType
  userId: string
  createdAt: string
}

export interface Expense {
  id: string
  amount: string
  note: string | null
  spentOn: string
  userId: string
  categoryId: string
  category?: Category
  createdAt: string
  updatedAt: string
}

export interface Income {
  id: string
  amount: string
  note: string | null
  receivedOn: string
  source: IncomeSource
  userId: string
  createdAt: string
  updatedAt: string
}

export interface BudgetSummary {
  income: string
  totalExpenses: string
  savingsTarget: string
  actualSavings: string
  remainingBudget: string
  targetMet: boolean
}

export interface CategorySlice {
  categoryId: string
  name: string
  type: CategoryType
  total: string
}

export interface TypeSlice {
  type: CategoryType
  total: string
}

export interface IncomeSlice {
  source: IncomeSource
  total: string
}

export interface Dashboard {
  month: string
  summary: BudgetSummary
  categoryBreakdown: CategorySlice[]
  typeBreakdown: TypeSlice[]
  incomeBreakdown: IncomeSlice[]
}

export interface TrendPoint {
  month: string
  income: string
  expenses: string
  net: string
}

export interface Trend {
  from: string
  to: string
  months: TrendPoint[]
}

export interface AppSetting {
  id: string
  savingsTargetType: SavingsTargetType
  savingsTargetValue: string
  updatedAt: string
  updatedBy: string | null
}

export const CATEGORY_TYPES: CategoryType[] = ['fixed', 'variable', 'wants', 'debts']

export const TYPE_LABELS: Record<CategoryType, string> = {
  fixed: 'Fixed',
  variable: 'Variable',
  wants: 'Wants',
  debts: 'Debts',
}

export const INCOME_SOURCES: IncomeSource[] = ['salary', 'side_project', 'other']

export const SOURCE_LABELS: Record<IncomeSource, string> = {
  salary: 'Salary',
  side_project: 'Side Project',
  other: 'Other',
}
