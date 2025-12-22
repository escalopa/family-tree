import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import { authApi } from '../api';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  setUser: (user: User | null) => void;
  isAuthenticated: boolean;
  hasRole: (minRole: number) => boolean;
  isActive: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      console.log('[AuthContext] Checking authentication...');
      try {
        // Try to fetch user from backend (will use cookies)
        // Backend middleware will auto-refresh tokens if needed
        console.log('[AuthContext] Calling GET /api/auth/me');
        const response = await authApi.getCurrentUser();
        console.log('[AuthContext] Auth successful, user:', response.user);
        setUser(response.user);
        localStorage.setItem('user', JSON.stringify(response.user));
      } catch (error: any) {
        // If 401, user is not authenticated (both tokens expired/invalid)
        console.log('[AuthContext] Auth failed:', error.response?.status, error.message);
        setUser(null);
        localStorage.removeItem('user');
      } finally {
        setLoading(false);
        console.log('[AuthContext] Loading complete');
      }
    };

    checkAuth();
  }, []);

  const updateUser = (newUser: User | null) => {
    setUser(newUser);
    if (newUser) {
      localStorage.setItem('user', JSON.stringify(newUser));
    } else {
      localStorage.removeItem('user');
    }
  };

  const hasRole = (minRole: number): boolean => {
    return user ? user.role_id >= minRole : false;
  };

  const value: AuthContextType = {
    user,
    loading,
    setUser: updateUser,
    isAuthenticated: !!user,
    hasRole,
    isActive: user?.is_active ?? false,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
