import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api } from '@/api/axios';
import { useAuthStore } from '@/stores/auth.store';
import { useAccountStore } from '@/stores/account.store';
import type { LoginRequest, RegisterRequest, AuthResponse } from '@/types/auth';

/**
 * Custom hook for authentication operations
 * Handles login, register, and logout with React Query
 */
export const useAuth = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { setAuth, clearAuth } = useAuthStore();
  const { clearAccount } = useAccountStore();

  // Login mutation
  const loginMutation = useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      const response = await api.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    },
    onSuccess: (data) => {
      // Save auth data to Zustand store (which persists to localStorage)
      setAuth(data.user, data.access_token, data.refresh_token);
      
      // Navigate to dashboard
      navigate('/dashboard');
    },
    onError: (error: any) => {
      console.error('Login error:', error);
      // Error handling is done in the component
    },
  });

  // Register mutation
  const registerMutation = useMutation({
    mutationFn: async (userData: RegisterRequest) => {
      const response = await api.post<AuthResponse>('/auth/register', userData);
      return response.data;
    },
    onSuccess: (data) => {
      // Save auth data to Zustand store (which persists to localStorage)
      setAuth(data.user, data.access_token, data.refresh_token);
      
      // Navigate to dashboard (or account creation page if needed)
      navigate('/dashboard');
    },
    onError: (error: any) => {
      console.error('Register error:', error);
      // Error handling is done in the component
    },
  });

  // Logout function
  const logout = () => {
    // Clear auth state
    clearAuth();
    
    // Clear account state
    clearAccount();
    
    // Clear all React Query cache
    queryClient.clear();
    
    // Navigate to login
    navigate('/login');
  };

  return {
    // Login
    login: loginMutation.mutate,
    loginAsync: loginMutation.mutateAsync,
    isLoggingIn: loginMutation.isPending,
    loginError: loginMutation.error,
    
    // Register
    register: registerMutation.mutate,
    registerAsync: registerMutation.mutateAsync,
    isRegistering: registerMutation.isPending,
    registerError: registerMutation.error,
    
    // Logout
    logout,
  };
};
