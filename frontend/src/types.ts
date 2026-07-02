export type Role = 'member' | 'admin'
export type CategoryType = 'fixed' | 'variable' | 'wants' | 'debts'
export type SavingsTargetType = 'percent' | 'fixed'

export interface User {
  id: string
  email: string
  role: Role
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

export interface Dashboard {
  month: string
  summary: BudgetSummary
  categoryBreakdown: CategorySlice[]
  typeBreakdown: TypeSlice[]
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
