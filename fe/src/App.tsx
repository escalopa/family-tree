import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';

// Pages
import LoginPage from './pages/LoginPage';
import CallbackPage from './pages/CallbackPage';
import InactivePage from './pages/InactivePage';
import UnauthorizedPage from './pages/UnauthorizedPage';
import TreePage from './pages/TreePage';
import LeaderboardPage from './pages/LeaderboardPage';
import MembersPage from './pages/MembersPage';
import UsersPage from './pages/UsersPage';
import UserProfilePage from './pages/UserProfilePage';

import { Roles } from './types';

const theme = createTheme({
  palette: {
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

const App: React.FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
        <Router>
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/auth/:provider/callback" element={<CallbackPage />} />
            <Route path="/inactive" element={<InactivePage />} />
            <Route path="/unauthorized" element={<UnauthorizedPage />} />

            {/* Protected Routes - Require Authentication */}
            <Route
              path="/tree"
              element={
                <ProtectedRoute requireActive>
                  <TreePage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/leaderboard"
              element={
                <ProtectedRoute requireActive>
                  <LeaderboardPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/users/:userId"
              element={
                <ProtectedRoute requireActive>
                  <UserProfilePage />
                </ProtectedRoute>
              }
            />

            {/* Admin Routes */}
            <Route
              path="/members"
              element={
                <ProtectedRoute requireActive minRole={Roles.ADMIN}>
                  <MembersPage />
                </ProtectedRoute>
              }
            />

            {/* Admin Routes - Can view users, only SuperAdmin can edit */}
            <Route
              path="/users"
              element={
                <ProtectedRoute requireActive minRole={Roles.ADMIN}>
                  <UsersPage />
                </ProtectedRoute>
              }
            />

            {/* Default Route */}
            <Route path="/" element={<Navigate to="/tree" replace />} />
            <Route path="*" element={<Navigate to="/tree" replace />} />
          </Routes>
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
};

export default App;
