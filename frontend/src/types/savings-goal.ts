import type { Currency } from './api';

export interface SavingsGoal {
  id: string;
  account_id: string;
  name: string;
  target_amount: number;
  current_amount: number;
  currency: Currency;
  deadline?: string | null; // YYYY-MM-DD
  created_at: string;
  updated_at: string;
}

export interface CreateSavingsGoalRequest {
  name: string;
  target_amount: number;
  currency: Currency;
  deadline?: string; // YYYY-MM-DD
}

export interface UpdateSavingsGoalRequest {
  name?: string;
  target_amount?: number;
  currency?: Currency;
  deadline?: string | ''; // "" para limpiar
}

export interface AddFundsRequest {
  amount: number;
  description?: string;
  date?: string; // YYYY-MM-DD, default: today
}

export interface WithdrawFundsRequest {
  amount: number;
  reason?: string;
  date?: string; // YYYY-MM-DD, default: today
}

export interface SavingsGoalTransaction {
  id: string;
  savings_goal_id: string;
  type: 'deposit' | 'withdrawal';
  amount: number;
  description?: string;
  date: string;
  created_at: string;
}
