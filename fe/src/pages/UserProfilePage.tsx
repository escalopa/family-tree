import React, { useEffect, useState, useMemo } from 'react';
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
  Button,
} from '@mui/material';
import '@mui/icons-material';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { usersApi } from '../api';
import { User, ScoreHistory, HistoryRecord } from '../types';
import { getRoleName, formatDateTime, formatRelativeTime } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';
import { Roles } from '../types';
import HistoryDiffDialog from '../components/HistoryDiffDialog';

const UserProfilePage: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { hasRole } = useAuth();
  const { getPreferredName, getAllNamesFormatted } = useLanguage();
  const [user, setUser] = useState<User | null>(null);
  const [scoreHistory, setScoreHistory] = useState<ScoreHistory[]>([]);
  const [displayedScoreCount, setDisplayedScoreCount] = useState(10);
  const [userChanges, setUserChanges] = useState<HistoryRecord[]>([]);
  const [displayedChangesCount, setDisplayedChangesCount] = useState(10);
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

  // Prepare data for the score history chart
  const chartData = useMemo(() => {
    if (!scoreHistory || scoreHistory.length === 0) return [];

    // Sort by date (oldest first)
    const sorted = [...scoreHistory].sort(
      (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    );

    // Calculate cumulative score with unique identifiers for each entry
    let cumulativeScore = 0;
    return sorted.map((score, index) => {
      cumulativeScore += score.points;
      const date = new Date(score.created_at);
      return {
        // Use index as unique key for X-axis to prevent grouping
        index: index,
        date: date.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
        }),
        time: date.toLocaleTimeString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
        }),
        dateTime: date.toLocaleString('en-US', {
          month: 'short',
          day: 'numeric',
          year: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
        }),
        points: score.points,
        cumulative: cumulativeScore,
        timestamp: date.getTime(),
        memberName: getPreferredName({ names: score.member_names }),
        fieldName: score.field_name,
      };
    });
  }, [scoreHistory, getPreferredName]);

  useEffect(() => {
    if (userId) {
      loadUserData();
    }
  }, [userId]);

  const loadUserData = async () => {
    try {
      const userResponse = await usersApi.getUser(Number(userId));
      setUser(userResponse);

      // Fetch ALL score history pages
      let allScores: ScoreHistory[] = [];
      let cursor: string | undefined = undefined;
      let hasMore = true;

      while (hasMore) {
        const scoresResponse = await usersApi.getScoreHistory(Number(userId), cursor);
        allScores = [...allScores, ...scoresResponse.scores];

        if (scoresResponse.next_cursor) {
          cursor = scoresResponse.next_cursor;
        } else {
          hasMore = false;
        }
      }

      setScoreHistory(allScores);
      setDisplayedScoreCount(10); // Reset pagination

      // Only super admins can see user changes
      if (hasRole(Roles.SUPER_ADMIN)) {
        // Fetch ALL user changes pages
        let allChanges: HistoryRecord[] = [];
        let changeCursor: string | undefined = undefined;
        let hasMoreChanges = true;

        while (hasMoreChanges) {
          const changesResponse = await usersApi.getUserChanges(Number(userId), changeCursor);
          allChanges = [...allChanges, ...changesResponse.history];

          if (changesResponse.next_cursor) {
            changeCursor = changesResponse.next_cursor;
          } else {
            hasMoreChanges = false;
          }
        }

        setUserChanges(allChanges);
        setDisplayedChangesCount(10); // Reset pagination
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
            {hasRole(Roles.SUPER_ADMIN) && <Tab label="Recent Changes" />}
          </Tabs>

          {/* Score History Tab */}
          {activeTab === 0 && (
            <Box sx={{ p: 2 }}>
              {/* Score History Chart */}
              {chartData.length > 0 && (
                <Box sx={{ mb: 4 }}>
                  <Typography variant="h6" gutterBottom>
                    Score Progress Over Time
                  </Typography>
                  <Box sx={{ mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                      Total Events: {chartData.length}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Final Cumulative Score: {chartData[chartData.length - 1]?.cumulative || 0}
                    </Typography>
                  </Box>
                  <ResponsiveContainer width="100%" height={350}>
                    <LineChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="index"
                        tick={{ fontSize: 12 }}
                        angle={-45}
                        textAnchor="end"
                        height={80}
                        tickFormatter={(index) => {
                          // Show date for every 5th point or if it's the first/last
                          const item = chartData[index];
                          if (index === 0 || index === chartData.length - 1 || index % 5 === 0) {
                            return item?.date || '';
                          }
                          return '';
                        }}
                        label={{ value: 'Timeline', position: 'insideBottom', offset: -5 }}
                      />
                      <YAxis
                        label={{ value: 'Points', angle: -90, position: 'insideLeft' }}
                        tick={{ fontSize: 12 }}
                      />
                      <Tooltip
                        content={({ active, payload }) => {
                          if (active && payload && payload.length > 0) {
                            const data = payload[0].payload;
                            return (
                              <Paper
                                sx={{
                                  p: 2,
                                  border: '1px solid #ccc',
                                  boxShadow: 2,
                                  minWidth: 250,
                                }}
                              >
                                <Typography variant="body2" fontWeight="bold" gutterBottom>
                                  {data.dateTime}
                                </Typography>
                                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                                  Member: {data.memberName}
                                </Typography>
                                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                                  Field: {data.fieldName}
                                </Typography>
                                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                    <Box
                                      sx={{
                                        width: 12,
                                        height: 12,
                                        backgroundColor: '#82ca9d',
                                        borderRadius: '50%',
                                      }}
                                    />
                                    <Typography variant="body2">
                                      Points Earned: <strong>+{data.points}</strong>
                                    </Typography>
                                  </Box>
                                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                    <Box
                                      sx={{
                                        width: 12,
                                        height: 12,
                                        backgroundColor: '#1976d2',
                                        borderRadius: '50%',
                                      }}
                                    />
                                    <Typography variant="body2">
                                      Cumulative Score: <strong>{data.cumulative}</strong>
                                    </Typography>
                                  </Box>
                                </Box>
                                <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                                  Event #{data.index + 1} of {chartData.length}
                                </Typography>
                              </Paper>
                            );
                          }
                          return null;
                        }}
                      />
                      <Legend />
                      <Line
                        type="monotone"
                        dataKey="cumulative"
                        stroke="#1976d2"
                        strokeWidth={2}
                        name="Cumulative Score"
                        dot={{ r: 4 }}
                        activeDot={{ r: 6 }}
                        connectNulls
                      />
                      <Line
                        type="monotone"
                        dataKey="points"
                        stroke="#82ca9d"
                        strokeWidth={2}
                        name="Points Earned"
                        dot={{ r: 3 }}
                        activeDot={{ r: 5 }}
                        connectNulls
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </Box>
              )}

              {/* Score History Table */}
              <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
                Detailed Score History
              </Typography>
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
                    {scoreHistory.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
                          <Typography variant="body2" color="text.secondary">
                            No score history available
                          </Typography>
                        </TableCell>
                      </TableRow>
                    ) : (
                      scoreHistory.slice(0, displayedScoreCount).map((score, idx) => (
                        <TableRow key={idx}>
                          <TableCell>
                            {getAllNamesFormatted({ names: score.member_names })}
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
                      ))
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
              {scoreHistory.length > displayedScoreCount && (
                <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                  <Button
                    variant="outlined"
                    onClick={() => setDisplayedScoreCount(prev => prev + 10)}
                  >
                    Load More ({scoreHistory.length - displayedScoreCount} remaining)
                  </Button>
                </Box>
              )}
            </Box>
          )}

          {/* Recent Changes Tab (Super Admin only) */}
          {activeTab === 1 && hasRole(Roles.SUPER_ADMIN) && (
            <Box sx={{ p: 2 }}>
              {userChanges.length === 0 ? (
                <Box sx={{ textAlign: 'center', py: 4 }}>
                  <Typography variant="body2" color="text.secondary">
                    No recent changes available
                  </Typography>
                </Box>
              ) : (
                <>
                  <TableContainer>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell>Change Type</TableCell>
                          <TableCell>Member ID</TableCell>
                          <TableCell>Date</TableCell>
                          <TableCell>Version</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {userChanges.slice(0, displayedChangesCount).map((change) => (
                          <TableRow
                            key={change.history_id}
                            hover
                            sx={{ cursor: 'pointer' }}
                            onClick={() => handleViewDiff(change)}
                          >
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
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                  {userChanges.length > displayedChangesCount && (
                    <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                      <Button
                        variant="outlined"
                        onClick={() => setDisplayedChangesCount(prev => prev + 10)}
                      >
                        Load More ({userChanges.length - displayedChangesCount} remaining)
                      </Button>
                    </Box>
                  )}
                </>
              )}
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
