import type { Currency, TransactionType } from './api';

export interface Income {
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
  income_type: TransactionType;
  date: string; // YYYY-MM-DD
  end_date?: string | null;
  recurring_income_id?: string | null;
  created_at: string;
}

// POST /incomes
export interface CreateIncomeRequest {
  description: string;
  amount: number;
  currency: Currency;
  date: string; // YYYY-MM-DD
  category_id?: string;
  family_member_id?: string;
  income_type?: TransactionType; // Default: "one-time"
  end_date?: string;
  // Multi-Currency Modo 3 (preferido)
  exchange_rate?: number;
  amount_in_primary_currency?: number;
}

// PUT /incomes/:id
export interface UpdateIncomeRequest {
  description?: string;
  amount?: number;
  currency?: Currency;
  date?: string;
  category_id?: string | '';
  family_member_id?: string | '';
  end_date?: string | '';
  exchange_rate?: number;
  amount_in_primary_currency?: number;
}

// GET /incomes - Query params
export interface IncomeListParams {
  month?: string; // YYYY-MM
  type?: TransactionType | 'all';
  category_id?: string;
  family_member_id?: string;
  currency?: Currency | 'all';
}

// GET /incomes - Response
export interface IncomeListResponse {
  incomes: Income[];
  count: number;
  summary: {
    total: number;
    byType: {
      'one-time': number;
      recurring: number;
    };
  };
}
