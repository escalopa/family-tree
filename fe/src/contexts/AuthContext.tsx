import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import { apiClient } from '../api';

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
    // Try to get current user on mount
    const checkAuth = async () => {
      try {
        // Check if we have a valid session by hitting a protected endpoint
        const response = await apiClient.get('/api/users/leaderboard');
        // If successful, we're authenticated, but we need to get user info from auth callback or storage
        const storedUser = localStorage.getItem('user');
        if (storedUser) {
          setUser(JSON.parse(storedUser));
        }
      } catch (error) {
        console.log('Not authenticated');
        setUser(null);
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
