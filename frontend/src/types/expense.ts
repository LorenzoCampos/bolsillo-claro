import type { Currency, TransactionType } from './api';

export interface Expense {
  id: string;
  account_id: string;
  family_member_id?: string | null;
  category_id?: string | null;
  category_name?: string | null;
  description: string;
  amount: number;
  currency: Currency;
  exchange_rate: number;
  amount_in_primary_currency: number;
  expense_type: TransactionType;
  date: string; // YYYY-MM-DD
  end_date?: string | null;
  recurring_expense_id?: string | null;
  created_at: string;
}

// POST /expenses - Request mínimo
export interface CreateExpenseRequest {
  description: string;
  amount: number;
  currency: Currency;
  date: string; // YYYY-MM-DD
  category_id?: string;
  family_member_id?: string;
  expense_type?: TransactionType; // Default: "one-time"
  end_date?: string;
  // Multi-Currency Modo 3 (preferido)
  exchange_rate?: number;
  amount_in_primary_currency?: number;
}

// PUT /expenses/:id - Partial update
export interface UpdateExpenseRequest {
  description?: string;
  amount?: number;
  currency?: Currency;
  date?: string;
  category_id?: string | ''; // "" para limpiar
  family_member_id?: string | '';
  end_date?: string | ''; // "" para limpiar
  exchange_rate?: number;
  amount_in_primary_currency?: number;
}

// GET /expenses - Query params
export interface ExpenseListParams {
  month?: string; // YYYY-MM
  type?: TransactionType | 'all';
  category_id?: string;
  family_member_id?: string;
  currency?: Currency | 'all';
}

// GET /expenses - Response
export interface ExpenseListResponse {
  expenses: Expense[];
  count: number;
  summary: {
    total: number;
    byType: {
      'one-time': number;
      recurring: number;
    };
  };
}
