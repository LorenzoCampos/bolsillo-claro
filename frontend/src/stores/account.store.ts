import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Account } from '@/types/account';

interface AccountState {
  activeAccountId: string | null;
  activeAccount: Account | null;
  
  // Actions
  setActiveAccount: (account: Account) => void;
  clearActiveAccount: () => void;
}

export const useAccountStore = create<AccountState>()(
  persist(
    (set) => ({
      activeAccountId: null,
      activeAccount: null,

      setActiveAccount: (account) => {
        // También guardar en localStorage para interceptors
        localStorage.setItem('active_account_id', account.id);
        
        set({
          activeAccountId: account.id,
          activeAccount: account,
        });
      },

      clearActiveAccount: () => {
        localStorage.removeItem('active_account_id');
        
        set({
          activeAccountId: null,
          activeAccount: null,
        });
      },
    }),
    {
      name: 'account-storage',
    }
  )
);
