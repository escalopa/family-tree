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
import { usersApi } from '../api';
import { User, Roles } from '../types';
import { getRoleName } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';

const UsersPage: React.FC = () => {
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
      <Box>
        <Typography variant="h4" gutterBottom>
          Users Management
        </Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          Manage user roles and access
        </Typography>

        {/* Filters */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6">
              Filters {!hasActiveFilters && '(Showing all users)'}
            </Typography>
            {hasActiveFilters && (
              <Button
                size="small"
                startIcon={<Clear />}
                onClick={handleClearFilters}
                color="secondary"
              >
                Clear Filters
              </Button>
            )}
          </Box>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Search by name or email"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type to search..."
              />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth>
                <InputLabel>Role</InputLabel>
                <Select
                  value={roleFilter}
                  label="Role"
                  onChange={(e) => setRoleFilter(e.target.value as number | 'all')}
                >
                  <MenuItem value="all">All Roles</MenuItem>
                  <MenuItem value={Roles.NONE}>None</MenuItem>
                  <MenuItem value={Roles.GUEST}>Guest</MenuItem>
                  <MenuItem value={Roles.ADMIN}>Admin</MenuItem>
                  <MenuItem value={Roles.SUPER_ADMIN}>Super Admin</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth>
                <InputLabel>Status</InputLabel>
                <Select
                  value={activeFilter}
                  label="Status"
                  onChange={(e) => setActiveFilter(e.target.value as boolean | 'all')}
                >
                  <MenuItem value="all">All Status</MenuItem>
                  <MenuItem value={true as any}>Active</MenuItem>
                  <MenuItem value={false as any}>Inactive</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </Grid>
        </Paper>

        {loading ? (
          <Typography>Loading...</Typography>
        ) : (
          <>
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Avatar</TableCell>
                    <TableCell>Name</TableCell>
                    <TableCell>Email</TableCell>
                    <TableCell>Role</TableCell>
                    <TableCell>Status</TableCell>
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
                      <TableCell>{user.full_name}</TableCell>
                      <TableCell>{user.email}</TableCell>
                      <TableCell>
                        <Chip
                          label={getRoleName(user.role_id)}
                          size="small"
                          color={user.role_id >= Roles.ADMIN ? 'primary' : 'default'}
                        />
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={user.is_active ? 'Active' : 'Inactive'}
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
                          No users found matching the filters
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
                  {loadingMore ? 'Loading...' : 'Load More'}
                </Button>
              </Box>
            )}
          </>
        )}

        {/* Edit User Dialog */}
        <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
          <DialogTitle>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h6">Edit User</Typography>
              {selectedUser && (
                <IconButton
                  onClick={() => handleOpenProfile(selectedUser.user_id)}
                  color="primary"
                  title="Open Profile"
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
                  <InputLabel>Role</InputLabel>
                  <Select
                    value={newRole}
                    label="Role"
                    onChange={(e) => setNewRole(Number(e.target.value))}
                  >
                    <MenuItem value={Roles.NONE}>None</MenuItem>
                    <MenuItem value={Roles.GUEST}>Guest</MenuItem>
                    <MenuItem value={Roles.ADMIN}>Admin</MenuItem>
                    <MenuItem value={Roles.SUPER_ADMIN}>Super Admin</MenuItem>
                  </Select>
                </FormControl>

                <FormControlLabel
                  control={
                    <Switch
                      checked={isActive}
                      onChange={(e) => setIsActive(e.target.checked)}
                    />
                  }
                  label="Active"
                  sx={{ mt: 2 }}
                />
              </Box>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseDialog}>Cancel</Button>
            <Button onClick={handleUpdateUser} variant="contained">
              Update
            </Button>
          </DialogActions>
        </Dialog>
      </Box>
    </Layout>
  );
};

export default UsersPage;
