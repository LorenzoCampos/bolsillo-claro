// ============================================
// TIPOS BASE DE LA API
// Basados en API.md v2.5
// ============================================

export type Currency = 'ARS' | 'USD' | 'EUR';
export type AccountType = 'personal' | 'family';
export type TransactionType = 'one-time' | 'recurring';
export type RecurrenceFrequency = 'daily' | 'weekly' | 'monthly' | 'yearly';

// Error response estándar
export interface ApiError {
  error: string;
  details?: string;
}

// Respuesta de paginación
export interface PaginatedResponse<T> {
  data: T[];
  count: number;
  total: number;
  page: number;
  limit: number;
}
