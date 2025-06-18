import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '@/types/auth';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
  setLoading: (loading: boolean) => void;
  initializeAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      setAuth: (user, token) => {
        localStorage.setItem('auth_token', token);
        set({ user, token, isAuthenticated: true });
      },
      clearAuth: () => {
        localStorage.removeItem('auth_token');
        set({ user: null, token: null, isAuthenticated: false });
      },
      setLoading: (loading) => set({ isLoading: loading }),
      initializeAuth: () => {
        const token = localStorage.getItem('auth_token');
        const state = get();
        
        console.log('🔍 Initializing auth state:', {
          tokenFromStorage: !!token,
          userFromState: !!state.user,
          isAuthenticated: state.isAuthenticated
        });
        
        if (token && state.user && !state.isAuthenticated) {
          // トークンとユーザー情報があるが認証状態がfalseの場合
          console.log('🔄 Restoring authentication state');
          set({ isAuthenticated: true, token });
        } else if (!token || !state.user) {
          // トークンまたはユーザー情報がない場合は認証状態をクリア
          console.log('❌ No valid auth data found, clearing state');
          set({ user: null, token: null, isAuthenticated: false });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ user: state.user, token: state.token, isAuthenticated: state.isAuthenticated }),
    }
  )
);