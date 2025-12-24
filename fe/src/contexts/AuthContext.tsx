import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import { authApi } from '../api';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  setUser: (user: User | null) => void;
  refreshUser: () => Promise<void>;
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
      try {
        // Try to fetch user from backend (will use cookies)
        // Backend middleware will auto-refresh tokens if needed
        const response = await authApi.getCurrentUser();
        setUser(response.user);
        localStorage.setItem('user', JSON.stringify(response.user));
      } catch (error: any) {
        // If 401, user is not authenticated (both tokens expired/invalid)
        setUser(null);
        localStorage.removeItem('user');
      } finally {
        setLoading(false);
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

  const refreshUser = async () => {
    try {
      const response = await authApi.getCurrentUser();
      setUser(response.user);
      localStorage.setItem('user', JSON.stringify(response.user));
    } catch (error: any) {
      setUser(null);
      localStorage.removeItem('user');
      throw error;
    }
  };

  const value: AuthContextType = {
    user,
    loading,
    setUser: updateUser,
    refreshUser,
    isAuthenticated: !!user,
    hasRole,
    isActive: user?.is_active ?? false,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
