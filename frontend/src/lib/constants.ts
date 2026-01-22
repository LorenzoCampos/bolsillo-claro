import type { Currency, AccountType, RecurrenceFrequency } from '@/types/api';

export const CURRENCIES: Currency[] = ['ARS', 'USD', 'EUR'];

export const CURRENCY_SYMBOLS: Record<Currency, string> = {
  ARS: '$',
  USD: 'US$',
  EUR: '€',
};

export const ACCOUNT_TYPES: AccountType[] = ['personal', 'family'];

export const ACCOUNT_TYPE_LABELS: Record<AccountType, string> = {
  personal: 'Personal',
  family: 'Familiar',
};

export const RECURRENCE_FREQUENCIES: RecurrenceFrequency[] = [
  'daily',
  'weekly',
  'monthly',
  'yearly',
];

export const RECURRENCE_FREQUENCY_LABELS: Record<RecurrenceFrequency, string> = {
  daily: 'Diario',
  weekly: 'Semanal',
  monthly: 'Mensual',
  yearly: 'Anual',
};

export const DAYS_OF_WEEK = [
  { value: 0, label: 'Domingo' },
  { value: 1, label: 'Lunes' },
  { value: 2, label: 'Martes' },
  { value: 3, label: 'Miércoles' },
  { value: 4, label: 'Jueves' },
  { value: 5, label: 'Viernes' },
  { value: 6, label: 'Sábado' },
];

export const API_DATE_FORMAT = 'yyyy-MM-dd'; // Para date-fns

// API Endpoints
export const API_ENDPOINTS = {
  AUTH: {
    REGISTER: '/auth/register',
    LOGIN: '/auth/login',
    REFRESH: '/auth/refresh',
  },
  ACCOUNTS: '/accounts',
  EXPENSES: '/expenses',
  INCOMES: '/incomes',
  RECURRING_EXPENSES: '/recurring-expenses',
  RECURRING_INCOMES: '/recurring-incomes',
  SAVINGS_GOALS: '/savings-goals',
  EXPENSE_CATEGORIES: '/expense-categories',
  INCOME_CATEGORIES: '/income-categories',
  DASHBOARD: {
    SUMMARY: '/dashboard/summary',
  },
} as const;
