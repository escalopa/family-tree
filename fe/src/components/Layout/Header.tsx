import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  IconButton,
  Avatar,
  Box,
} from '@mui/material';
import {
  AccountTree,
  People,
  SupervisorAccount,
  Leaderboard,
  AccountCircle,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { Roles } from '../../types';

const Header: React.FC = () => {
  const navigate = useNavigate();
  const { user, hasRole, isActive } = useAuth();

  return (
    <AppBar position="static">
      <Toolbar>
        <AccountTree sx={{ mr: 2 }} />
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          Family Tree
        </Typography>

        {user && isActive && (
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              color="inherit"
              startIcon={<AccountTree />}
              onClick={() => navigate('/tree')}
            >
              Tree
            </Button>
            <Button
              color="inherit"
              startIcon={<Leaderboard />}
              onClick={() => navigate('/leaderboard')}
            >
              Leaderboard
            </Button>
            {hasRole(Roles.ADMIN) && (
              <Button
                color="inherit"
                startIcon={<People />}
                onClick={() => navigate('/members')}
              >
                Members
              </Button>
            )}
            {hasRole(Roles.SUPER_ADMIN) && (
              <Button
                color="inherit"
                startIcon={<SupervisorAccount />}
                onClick={() => navigate('/users')}
              >
                Users
              </Button>
            )}
          </Box>
        )}

        {user ? (
          <IconButton
            size="large"
            aria-label="profile"
            onClick={() => navigate('/profile')}
            color="inherit"
            sx={{ ml: 2 }}
          >
            {user.avatar ? (
              <Avatar src={user.avatar} alt={user.full_name} />
            ) : (
              <AccountCircle />
            )}
          </IconButton>
        ) : (
          <Button color="inherit" onClick={() => navigate('/login')}>
            Login
          </Button>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Header;
