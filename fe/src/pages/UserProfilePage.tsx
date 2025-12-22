import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  Box,
  Paper,
  Typography,
  Avatar,
  Grid,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  Tab,
  IconButton,
} from '@mui/material';
import { Visibility } from '@mui/icons-material';
import { usersApi } from '../api';
import { User, ScoreHistory, HistoryRecord } from '../types';
import { getRoleName, formatDate, formatDateTime, formatRelativeTime } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';
import { Roles } from '../types';
import HistoryDiffDialog from '../components/HistoryDiffDialog';

const UserProfilePage: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { hasRole } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [scoreHistory, setScoreHistory] = useState<ScoreHistory[]>([]);
  const [userChanges, setUserChanges] = useState<HistoryRecord[]>([]);
  const [activeTab, setActiveTab] = useState(0);
  const [loading, setLoading] = useState(true);
  const [selectedHistory, setSelectedHistory] = useState<HistoryRecord | null>(null);
  const [diffDialogOpen, setDiffDialogOpen] = useState(false);

  const handleViewDiff = (history: HistoryRecord) => {
    setSelectedHistory(history);
    setDiffDialogOpen(true);
  };

  const handleCloseDiff = () => {
    setDiffDialogOpen(false);
    setSelectedHistory(null);
  };

  useEffect(() => {
    if (userId) {
      loadUserData();
    }
  }, [userId]);

  const loadUserData = async () => {
    try {
      const userResponse = await usersApi.getUser(Number(userId));
      setUser(userResponse);

      const scoresResponse = await usersApi.getScoreHistory(Number(userId));
      setScoreHistory(scoresResponse.scores);

      // Admins and super admins can see user changes
      if (hasRole(Roles.ADMIN)) {
        const changesResponse = await usersApi.getUserChanges(Number(userId));
        setUserChanges(changesResponse.history);
      }
    } catch (error) {
      console.error('Failed to load user data:', error);
    } finally {
      setLoading(false);
    }
  };


  if (loading) {
    return (
      <Layout>
        <Typography>Loading...</Typography>
      </Layout>
    );
  }

  if (!user) {
    return (
      <Layout>
        <Typography>User not found</Typography>
      </Layout>
    );
  }

  return (
    <Layout>
      <Box>
        {/* User Info Card */}
        <Paper sx={{ p: 3, mb: 3 }}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={4} sx={{ textAlign: 'center' }}>
              <Avatar
                src={user.avatar || undefined}
                sx={{ width: 150, height: 150, mx: 'auto', mb: 2 }}
              >
                {user.full_name[0]}
              </Avatar>
              <Typography variant="h5" gutterBottom>
                {user.full_name}
              </Typography>
              <Typography variant="body1" color="text.secondary" gutterBottom>
                {user.email}
              </Typography>
              <Box sx={{ mt: 2, display: 'flex', gap: 1, justifyContent: 'center' }}>
                <Chip
                  label={getRoleName(user.role_id)}
                  color={user.role_id >= Roles.ADMIN ? 'primary' : 'default'}
                />
                <Chip
                  label={user.is_active ? 'Active' : 'Inactive'}
                  color={user.is_active ? 'success' : 'default'}
                />
              </Box>
            </Grid>
            <Grid item xs={12} md={8}>
              <Typography variant="h6" gutterBottom>
                Statistics
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="h4" color="primary">
                      {user.total_score || 0}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Total Score
                    </Typography>
                  </Paper>
                </Grid>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="h4" color="primary">
                      {scoreHistory.length}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Contributions
                    </Typography>
                  </Paper>
                </Grid>
              </Grid>
            </Grid>
          </Grid>
        </Paper>

        {/* Tabs */}
        <Paper>
          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)}>
            <Tab label="Score History" />
            {hasRole(Roles.ADMIN) && <Tab label="Recent Changes" />}
          </Tabs>

          {/* Score History Tab */}
          {activeTab === 0 && (
            <Box sx={{ p: 2 }}>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Member</TableCell>
                      <TableCell>Field</TableCell>
                      <TableCell>Points</TableCell>
                      <TableCell>Date</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {scoreHistory.map((score, idx) => (
                      <TableRow key={idx}>
                        <TableCell>
                          {score.member_arabic_name} ({score.member_english_name})
                        </TableCell>
                        <TableCell>{score.field_name}</TableCell>
                        <TableCell>
                          <Chip label={`+${score.points}`} color="primary" size="small" />
                        </TableCell>
                        <TableCell>
                          <Box>
                            <Typography variant="body2">{formatDateTime(score.created_at)}</Typography>
                            <Typography variant="caption" color="text.secondary">
                              {formatRelativeTime(score.created_at)}
                            </Typography>
                          </Box>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>
          )}

          {/* Recent Changes Tab (Admin and Super Admin) */}
          {activeTab === 1 && hasRole(Roles.ADMIN) && (
            <Box sx={{ p: 2 }}>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Change Type</TableCell>
                      <TableCell>Member ID</TableCell>
                      <TableCell>Date</TableCell>
                      <TableCell>Version</TableCell>
                      <TableCell>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {userChanges.map((change) => (
                      <TableRow key={change.history_id} hover>
                        <TableCell>
                          <Chip label={change.change_type} size="small" />
                        </TableCell>
                        <TableCell>{change.member_id}</TableCell>
                        <TableCell>
                          <Box>
                            <Typography variant="body2">{formatDateTime(change.changed_at)}</Typography>
                            <Typography variant="caption" color="text.secondary">
                              {formatRelativeTime(change.changed_at)}
                            </Typography>
                          </Box>
                        </TableCell>
                        <TableCell>{change.member_version}</TableCell>
                        <TableCell>
                          <IconButton
                            size="small"
                            onClick={() => handleViewDiff(change)}
                            color="primary"
                          >
                            <Visibility />
                          </IconButton>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>
          )}
        </Paper>

        {/* History Diff Dialog */}
        <HistoryDiffDialog
          open={diffDialogOpen}
          onClose={handleCloseDiff}
          history={selectedHistory}
        />
      </Box>
    </Layout>
  );
};

export default UserProfilePage;
