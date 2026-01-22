export interface ExpenseCategory {
  id: string;
  account_id?: string | null; // NULL si es system
  name: string;
  icon?: string | null;
  color?: string | null;
  is_system: boolean;
  created_at: string;
}

export interface IncomeCategory {
  id: string;
  account_id?: string | null;
  name: string;
  icon?: string | null;
  color?: string | null;
  is_system: boolean;
  created_at: string;
}

export interface CreateCategoryRequest {
  name: string;
  icon?: string;
  color?: string;
}

export interface UpdateCategoryRequest {
  name?: string;
  icon?: string;
  color?: string;
}

export interface CategoryListResponse<T> {
  categories: T[];
  count: number;
}
