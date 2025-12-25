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
  Groups,
  AdminPanelSettings,
  Leaderboard,
  AccountCircle,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../../contexts/AuthContext';
import { Roles } from '../../types';

const Header: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user, hasRole, isActive } = useAuth();

  return (
    <AppBar position="static">
      <Toolbar>
        <AccountTree sx={{ mr: 2 }} />
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          {t('navigation.familyTree')}
        </Typography>

        {user && isActive && (
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              color="inherit"
              startIcon={<AccountTree />}
              onClick={() => navigate('/tree')}
            >
              {t('navigation.tree')}
            </Button>
            <Button
              color="inherit"
              startIcon={<Leaderboard />}
              onClick={() => navigate('/leaderboard')}
            >
              {t('navigation.leaderboard')}
            </Button>
            {hasRole(Roles.ADMIN) && (
              <Button
                color="inherit"
                startIcon={<Groups />}
                onClick={() => navigate('/members')}
              >
                {t('navigation.members')}
              </Button>
            )}
            {hasRole(Roles.SUPER_ADMIN) && (
              <Button
                color="inherit"
                startIcon={<AdminPanelSettings />}
                onClick={() => navigate('/users')}
              >
                {t('navigation.users')}
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
            {t('common.login')}
          </Button>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Header;
