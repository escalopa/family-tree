import React from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Container,
  Typography,
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Avatar,
  IconButton,
} from '@mui/material';
import { Layout } from '../components/common/Layout';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { usersApi } from '../api/users';
import { ROLE_LABELS } from '../utils/constants';
import EditIcon from '@mui/icons-material/Edit';

export const UsersPage: React.FC = () => {
  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => usersApi.listUsers(),
  });

  if (isLoading) {
    return (
      <Layout>
        <LoadingSpinner />
      </Layout>
    );
  }

  return (
    <Layout>
      <Container maxWidth="lg">
        <Typography variant="h4" gutterBottom>
          Manage Users
        </Typography>

        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>User</TableCell>
                <TableCell>Email</TableCell>
                <TableCell>Role</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {users?.map((user) => (
                <TableRow key={user.user_id}>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <Avatar src={user.avatar || undefined} alt={user.full_name} />
                      <Typography>{user.full_name}</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <Chip
                      label={ROLE_LABELS[user.role_id as keyof typeof ROLE_LABELS]}
                      size="small"
                      color={user.role_id >= 300 ? 'primary' : 'default'}
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
                    <IconButton size="small">
                      <EditIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Container>
    </Layout>
  );
};



