import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Box, CircularProgress } from '@mui/material';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireActive?: boolean;
  minRole?: number;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requireActive = false,
  minRole,
}) => {
  const { user, loading, isActive, hasRole } = useAuth();

  if (loading) {
    return (
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (requireActive && !isActive) {
    return <Navigate to="/inactive" replace />;
  }

  if (minRole !== undefined && !hasRole(minRole)) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <>{children}</>;
};

export default ProtectedRoute;
