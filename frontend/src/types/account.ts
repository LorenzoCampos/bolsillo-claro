import type { Currency, AccountType } from './api';

export interface Account {
  id: string;
  user_id: string;
  name: string;
  type: AccountType;
  currency: Currency;
  created_at: string;
  updated_at: string;
}

export interface CreateAccountRequest {
  name: string;
  type: AccountType;
  currency: Currency;
}

export interface UpdateAccountRequest {
  name?: string;
  currency?: Currency;
}

export interface FamilyMember {
  id: string;
  account_id: string;
  name: string;
  is_active: boolean;
  created_at: string;
}

export interface CreateFamilyMemberRequest {
  name: string;
}

export interface UpdateFamilyMemberRequest {
  name?: string;
}
