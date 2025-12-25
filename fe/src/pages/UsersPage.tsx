import React, { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Avatar,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  TextField,
  Grid,
} from '@mui/material';
import { OpenInNew, Clear } from '@mui/icons-material';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { usersApi } from '../api';
import { User, Roles } from '../types';
import { getRoleName } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';

const UsersPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { hasRole } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [nextCursor, setNextCursor] = useState<string | undefined>(undefined);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [newRole, setNewRole] = useState<number>(Roles.NONE);
  const [isActive, setIsActive] = useState(false);
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);

  // Filter states from URL query params
  const [searchQuery, setSearchQuery] = useState(() => searchParams.get('search') || '');
  const [roleFilter, setRoleFilter] = useState<number | 'all'>(() => {
    const role = searchParams.get('role');
    return role && role !== 'all' ? Number(role) : 'all';
  });
  const [activeFilter, setActiveFilter] = useState<boolean | 'all'>(() => {
    const active = searchParams.get('active');
    if (active === 'true') return true;
    if (active === 'false') return false;
    return 'all';
  });

  useEffect(() => {
    // Update URL params when filters change
    const params = new URLSearchParams();
    if (searchQuery) params.set('search', searchQuery);
    if (roleFilter !== 'all') params.set('role', roleFilter.toString());
    if (activeFilter !== 'all') params.set('active', activeFilter.toString());
    setSearchParams(params, { replace: true });

    loadUsers();
  }, [searchQuery, roleFilter, activeFilter]);

  const loadUsers = async (cursor?: string) => {
    try {
      const isLoadingMore = !!cursor;
      if (isLoadingMore) {
        setLoadingMore(true);
      } else {
        setLoading(true);
      }

      // Convert filter values for API
      const roleIdFilter = roleFilter === 'all' ? undefined : roleFilter;
      const activeFilterValue = activeFilter === 'all' ? undefined : activeFilter;

      const response = await usersApi.listUsers(
        cursor,
        20,
        searchQuery || undefined,
        roleIdFilter,
        activeFilterValue
      );

      if (isLoadingMore) {
        setUsers((prev) => [...prev, ...(response.users || [])]);
      } else {
        setUsers(response.users || []);
      }

      setNextCursor(response.next_cursor);
    } catch (error) {
      console.error('load users:', error);
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  const loadMoreUsers = () => {
    if (nextCursor && !loadingMore) {
      loadUsers(nextCursor);
    }
  };

  const handleOpenDialog = (user: User) => {
    setSelectedUser(user);
    setNewRole(user.role_id);
    setIsActive(user.is_active);
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedUser(null);
  };

  const handleUpdateUser = async () => {
    if (!selectedUser) return;

    try {
      // Update role if changed
      if (newRole !== selectedUser.role_id) {
        await usersApi.updateRole(selectedUser.user_id, { role_id: newRole });
      }

      // Update active status if changed
      if (isActive !== selectedUser.is_active) {
        await usersApi.updateActive(selectedUser.user_id, { is_active: isActive });
      }

      handleCloseDialog();
      loadUsers(); // Refresh list
    } catch (error) {
      console.error('update user:', error);
    }
  };

  const handleOpenProfile = (userId: number) => {
    navigate(`/users/${userId}`);
  };

  const handleClearFilters = () => {
    setSearchQuery('');
    setRoleFilter('all');
    setActiveFilter('all');
  };

  const hasActiveFilters = searchQuery || roleFilter !== 'all' || activeFilter !== 'all';


  return (
    <Layout>
      <Box sx={{ width: '100%' }}>
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
        <Typography variant="h4" gutterBottom sx={{ mb: 3 }}>
          {t('users.management')}
        </Typography>
        </motion.div>

        {/* Filters */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
        >
        <Paper sx={{ p: 2, mb: 3 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6">
              {t('common.filter')} {!hasActiveFilters && t('users.showingAllUsers')}
            </Typography>
            {hasActiveFilters && (
              <Button
                size="small"
                startIcon={<Clear />}
                onClick={handleClearFilters}
                color="secondary"
              >
                {t('tree.clearFilters')}
              </Button>
            )}
          </Box>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label={t('users.searchByNameOrEmail')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('users.typeToSearch')}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth>
                <InputLabel>{t('user.role')}</InputLabel>
                <Select
                  value={roleFilter}
                  label={t('user.role')}
                  onChange={(e) => setRoleFilter(e.target.value as number | 'all')}
                >
                  <MenuItem value="all">{t('users.allRoles')}</MenuItem>
                  <MenuItem value={Roles.NONE}>{t('roles.none')}</MenuItem>
                  <MenuItem value={Roles.GUEST}>{t('roles.guest')}</MenuItem>
                  <MenuItem value={Roles.ADMIN}>{t('roles.admin')}</MenuItem>
                  <MenuItem value={Roles.SUPER_ADMIN}>{t('roles.superAdmin')}</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth>
                <InputLabel>{t('user.status')}</InputLabel>
                <Select
                  value={activeFilter}
                  label={t('user.status')}
                  onChange={(e) => setActiveFilter(e.target.value as boolean | 'all')}
                >
                  <MenuItem value="all">{t('users.allStatus')}</MenuItem>
                  <MenuItem value={true as any}>{t('user.active')}</MenuItem>
                  <MenuItem value={false as any}>{t('user.inactive')}</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </Grid>
        </Paper>
        </motion.div>

        {loading ? (
          <Typography>{t('common.loading')}</Typography>
        ) : (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.2 }}
          >
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell className="table-header-cell">{t('member.avatar')}</TableCell>
                    <TableCell className="table-header-cell">{t('member.name')}</TableCell>
                    <TableCell className="table-header-cell email-cell">{t('user.email')}</TableCell>
                    <TableCell className="table-header-cell">{t('user.role')}</TableCell>
                    <TableCell className="table-header-cell">{t('user.status')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {users.map((user) => (
                    <TableRow
                      key={user.user_id}
                      hover
                      sx={{ cursor: isSuperAdmin ? 'pointer' : 'default' }}
                      onClick={() => isSuperAdmin && handleOpenDialog(user)}
                    >
                      <TableCell>
                        <Avatar src={user.avatar || undefined}>{user.full_name[0]}</Avatar>
                      </TableCell>
                      <TableCell className="mixed-content-cell">{user.full_name}</TableCell>
                      <TableCell className="email-cell">{user.email}</TableCell>
                      <TableCell>
                        <Chip
                          label={getRoleName(user.role_id, t)}
                          size="small"
                          color={user.role_id >= Roles.ADMIN ? 'primary' : 'default'}
                        />
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={user.is_active ? t('user.active') : t('user.inactive')}
                          size="small"
                          color={user.is_active ? 'success' : 'default'}
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                  {users.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                        <Typography variant="body2" color="text.secondary">
                          {t('users.noUsersMatchingFilters')}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>

            {nextCursor && (
              <Box sx={{ mt: 2, textAlign: 'center' }}>
                <Button
                  variant="outlined"
                  onClick={loadMoreUsers}
                  disabled={loadingMore}
                >
                  {loadingMore ? t('common.loading') : t('users.loadMore')}
                </Button>
              </Box>
            )}
          </motion.div>
        )}

        {/* Edit User Dialog */}
        <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
          <DialogTitle>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h6">{t('users.editUser')}</Typography>
              {selectedUser && (
                <IconButton
                  onClick={() => handleOpenProfile(selectedUser.user_id)}
                  color="primary"
                  title={t('users.openProfile')}
                >
                  <OpenInNew />
                </IconButton>
              )}
            </Box>
          </DialogTitle>
          <DialogContent>
            {selectedUser && (
              <Box sx={{ mt: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
                  <Avatar src={selectedUser.avatar || undefined} sx={{ width: 56, height: 56 }}>
                    {selectedUser.full_name[0]}
                  </Avatar>
                  <Box>
                    <Typography variant="subtitle1">
                      {selectedUser.full_name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {selectedUser.email}
                    </Typography>
                  </Box>
                </Box>

                <FormControl fullWidth sx={{ mt: 3 }}>
                  <InputLabel>{t('user.role')}</InputLabel>
                  <Select
                    value={newRole}
                    label={t('user.role')}
                    onChange={(e) => setNewRole(Number(e.target.value))}
                  >
                    <MenuItem value={Roles.NONE}>{t('roles.none')}</MenuItem>
                    <MenuItem value={Roles.GUEST}>{t('roles.guest')}</MenuItem>
                    <MenuItem value={Roles.ADMIN}>{t('roles.admin')}</MenuItem>
                    <MenuItem value={Roles.SUPER_ADMIN}>{t('roles.superAdmin')}</MenuItem>
                  </Select>
                </FormControl>

                <FormControlLabel
                  control={
                    <Switch
                      checked={isActive}
                      onChange={(e) => setIsActive(e.target.checked)}
                    />
                  }
                  label={t('user.active')}
                  sx={{ mt: 2 }}
                />
              </Box>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseDialog}>{t('common.cancel')}</Button>
            <Button onClick={handleUpdateUser} variant="contained">
              {t('users.update')}
            </Button>
          </DialogActions>
        </Dialog>
      </Box>
    </Layout>
  );
};

export default UsersPage;
