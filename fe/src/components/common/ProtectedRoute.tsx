import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { ROLES } from '../../types/user';

interface ProtectedRouteProps {
  children: React.ReactNode;
  minRole?: number;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, minRole = ROLES.GUEST }) => {
  const { user, isAuthenticated } = useAuthStore();

  if (!isAuthenticated || !user) {
    return <Navigate to="/login" replace />;
  }

  if (user.role_id === ROLES.NONE) {
    return <Navigate to="/none" replace />;
  }

  if (minRole && user.role_id < minRole) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};


