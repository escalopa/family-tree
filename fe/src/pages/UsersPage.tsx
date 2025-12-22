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
} from '@mui/material';
import { Edit, Visibility } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { usersApi } from '../api';
import { User, Roles } from '../types';
import { getRoleName } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';

const UsersPage: React.FC = () => {
  const navigate = useNavigate();
  const { hasRole } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [nextCursor, setNextCursor] = useState<string | undefined>(undefined);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [newRole, setNewRole] = useState<number>(Roles.NONE);
  const [isActive, setIsActive] = useState(false);
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async (cursor?: string) => {
    try {
      const isLoadingMore = !!cursor;
      if (isLoadingMore) {
        setLoadingMore(true);
      }

      const response = await usersApi.listUsers(cursor, 20);

      if (isLoadingMore) {
        setUsers((prev) => [...prev, ...response.users]);
      } else {
        setUsers(response.users);
      }

      setNextCursor(response.next_cursor);
    } catch (error) {
      console.error('Failed to load users:', error);
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
      console.error('Failed to update user:', error);
    }
  };

  return (
    <Layout>
      <Box>
        <Typography variant="h4" gutterBottom>
          Users Management
        </Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          Manage user roles and access
        </Typography>

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
                    <TableCell>Total Score</TableCell>
                    <TableCell>Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {users.map((user) => (
                    <TableRow key={user.user_id}>
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
                      <TableCell>
                        {user.total_score !== undefined ? user.total_score : '-'}
                      </TableCell>
                      <TableCell>
                        <IconButton
                          size="small"
                          onClick={() => navigate(`/users/${user.user_id}`)}
                        >
                          <Visibility />
                        </IconButton>
                        {isSuperAdmin && (
                          <IconButton size="small" onClick={() => handleOpenDialog(user)}>
                            <Edit />
                          </IconButton>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
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
          <DialogTitle>Edit User</DialogTitle>
          <DialogContent>
            {selectedUser && (
              <Box sx={{ mt: 2 }}>
                <Typography variant="subtitle1" gutterBottom>
                  {selectedUser.full_name}
                </Typography>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  {selectedUser.email}
                </Typography>

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
