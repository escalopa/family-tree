import React, { useEffect, useState, useMemo } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
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
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { usersApi } from '../api';
import { User, ScoreHistory, HistoryRecord } from '../types';
import { getRoleName, formatDateTime, formatRelativeTime, getChangeTypeColor } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';
import { Roles } from '../types';
import HistoryDiffDialog from '../components/HistoryDiffDialog';

const UserProfilePage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { userId } = useParams<{ userId: string }>();
  const { hasRole } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const [user, setUser] = useState<User | null>(null);
  const [scoreHistory, setScoreHistory] = useState<ScoreHistory[]>([]);
  const [displayedScoreCount, setDisplayedScoreCount] = useState(10);
  const [userChanges, setUserChanges] = useState<HistoryRecord[]>([]);
  const [displayedChangesCount, setDisplayedChangesCount] = useState(10);
  const [activeTab, setActiveTab] = useState(() => {
    const tab = searchParams.get('tab');
    return tab === 'changes' && hasRole(Roles.SUPER_ADMIN) ? 1 : 0;
  });
  const [loading, setLoading] = useState(true);
  const [loadingScoreHistory, setLoadingScoreHistory] = useState(false);
  const [loadingUserChanges, setLoadingUserChanges] = useState(false);
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

    // Get current locale from i18n
    const currentLocale = i18n.language || 'en';

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
        date: date.toLocaleDateString(currentLocale, {
          month: 'short',
          day: 'numeric',
        }),
        time: date.toLocaleTimeString(currentLocale, {
          hour: '2-digit',
          minute: '2-digit',
        }),
        dateTime: date.toLocaleString(currentLocale, {
          month: 'short',
          day: 'numeric',
          year: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
        }),
        points: score.points,
        cumulative: cumulativeScore,
        timestamp: date.getTime(),
        memberName: score.member_name,
        fieldName: score.field_name,
      };
    });
  }, [scoreHistory, i18n.language]);

  useEffect(() => {
    if (userId) {
      loadUserData();
    }
  }, [userId]);

  useEffect(() => {
    // Load data based on active tab
    if (userId && user) {
      if (activeTab === 0 && scoreHistory.length === 0) {
        loadScoreHistory();
      } else if (activeTab === 1 && hasRole(Roles.SUPER_ADMIN) && userChanges.length === 0) {
        loadUserChanges();
      }
    }
  }, [activeTab, userId, user]);

  const loadUserData = async () => {
    try {
      const userResponse = await usersApi.getUser(Number(userId));
      setUser(userResponse);
    } catch (error) {

    } finally {
      setLoading(false);
    }
  };

  const loadScoreHistory = async () => {
    if (!userId) return;

    setLoadingScoreHistory(true);
    try {
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
    } catch (error) {

    } finally {
      setLoadingScoreHistory(false);
    }
  };

  const loadUserChanges = async () => {
    if (!userId || !hasRole(Roles.SUPER_ADMIN)) return;

    setLoadingUserChanges(true);
    try {
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
    } catch (error) {

    } finally {
      setLoadingUserChanges(false);
    }
  };


  if (loading) {
    return (
      <Layout>
        <Typography>{t('common.loading')}</Typography>
      </Layout>
    );
  }

  if (!user) {
    return (
      <Layout>
        <Typography>{t('userProfile.userNotFound')}</Typography>
      </Layout>
    );
  }

  return (
    <Layout>
      <Box sx={{ width: '100%' }}>
        {/* User Info Card */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
        <Paper sx={{ p: 3, mb: 3 }}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={4} sx={{ textAlign: 'center', display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center' }}>
              <Avatar
                src={user.avatar || undefined}
                sx={{ width: 180, height: 180, mb: 2 }}
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
                  label={getRoleName(user.role_id, t)}
                  color={user.role_id >= Roles.ADMIN ? 'primary' : 'default'}
                />
                <Chip
                  label={user.is_active ? t('user.active') : t('user.inactive')}
                  color={user.is_active ? 'success' : 'default'}
                />
              </Box>
            </Grid>
            <Grid item xs={12} md={8}>
              <Typography variant="h6" gutterBottom>
                {t('userProfile.statistics')}
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="h4" color="primary">
                      {user.total_score || 0}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {t('user.totalScore')}
                    </Typography>
                  </Paper>
                </Grid>
                <Grid item xs={6}>
                  <Paper sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="h4" color="primary">
                      {scoreHistory.length}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {t('userProfile.contributions')}
                    </Typography>
                  </Paper>
                </Grid>
              </Grid>
            </Grid>
          </Grid>
        </Paper>
        </motion.div>

        {/* Tabs */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
        >
        <Paper>
          <Tabs
            value={activeTab}
            onChange={(_, v) => {
              setActiveTab(v);
              const params = new URLSearchParams(searchParams);
              if (v === 0) {
                params.delete('tab');
              } else {
                params.set('tab', 'changes');
              }
              setSearchParams(params, { replace: true });
            }}
          >
            <Tab label={t('userProfile.scoreHistory')} />
            {hasRole(Roles.SUPER_ADMIN) && <Tab label={t('userProfile.recentChanges')} />}
          </Tabs>

          {/* Score History Tab */}
          {activeTab === 0 && (
            <Box sx={{ p: 2 }}>
              {loadingScoreHistory ? (
                <Box sx={{ textAlign: 'center', py: 4 }}>
                  <Typography>{t('userProfile.loadingScoreHistory')}</Typography>
                </Box>
              ) : (
                <>
              {/* Score History Chart */}
              {chartData.length > 0 && (
                <Box sx={{ mb: 4 }}>
                  <Typography variant="h6" gutterBottom>
                    {t('userProfile.scoreProgressOverTime')}
                  </Typography>
                  <Box sx={{ mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                      {t('userProfile.totalEvents')}: {chartData.length}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {t('userProfile.finalCumulativeScore')}: {chartData[chartData.length - 1]?.cumulative || 0}
                    </Typography>
                  </Box>
                  <ResponsiveContainer width="100%" height={420}>
                    <LineChart data={chartData} margin={{ top: 30, right: 30, left: 20, bottom: 70 }}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="index"
                        tick={{ fontSize: 12 }}
                        angle={-45}
                        textAnchor="end"
                        height={90}
                        tickFormatter={(index) => {
                          // Show date for every 5th point or if it's the first/last
                          const item = chartData[index];
                          if (index === 0 || index === chartData.length - 1 || index % 5 === 0) {
                            return item?.date || '';
                          }
                          return '';
                        }}
                        label={{ value: t('userProfile.timeline'), position: 'insideBottom', offset: -25 }}
                      />
                      <YAxis
                        label={{ value: t('userProfile.points'), angle: -90, position: 'insideLeft' }}
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
                                  {t('userProfile.member')}: {data.memberName}
                                </Typography>
                                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                                  {t('userProfile.field')}: {data.fieldName}
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
                                      {t('userProfile.pointsEarned')}: <strong>+{data.points}</strong>
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
                                      {t('userProfile.cumulativeScore')}: <strong>{data.cumulative}</strong>
                                    </Typography>
                                  </Box>
                                </Box>
                                <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                                  {t('userProfile.event')} #{data.index + 1} {t('userProfile.of')} {chartData.length}
                                </Typography>
                              </Paper>
                            );
                          }
                          return null;
                        }}
                      />
                      <Legend verticalAlign="top" height={36} />
                      <Line
                        type="monotone"
                        dataKey="cumulative"
                        stroke="#1976d2"
                        strokeWidth={2}
                        name={t('userProfile.cumulativeScore')}
                        dot={{ r: 4 }}
                        activeDot={{ r: 6 }}
                        connectNulls
                      />
                      <Line
                        type="monotone"
                        dataKey="points"
                        stroke="#82ca9d"
                        strokeWidth={2}
                        name={t('userProfile.pointsEarned')}
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
                {t('userProfile.detailedScoreHistory')}
              </Typography>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell className="table-header-cell">{t('userProfile.member')}</TableCell>
                      <TableCell className="table-header-cell">{t('userProfile.field')}</TableCell>
                      <TableCell className="table-header-cell numeric-cell">{t('userProfile.points')}</TableCell>
                      <TableCell className="table-header-cell numeric-cell">{t('userProfile.date')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {scoreHistory.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
                          <Typography variant="body2" color="text.secondary">
                            {t('userProfile.noScoreHistory')}
                          </Typography>
                        </TableCell>
                      </TableRow>
                    ) : (
                      scoreHistory.slice(0, displayedScoreCount).map((score, idx) => (
                        <TableRow key={idx}>
                          <TableCell className="mixed-content-cell">
                            {score.member_name}
                          </TableCell>
                          <TableCell>{score.field_name}</TableCell>
                          <TableCell className="numeric-cell">
                            <Chip label={`+${score.points}`} color="primary" size="small" />
                          </TableCell>
                          <TableCell className="numeric-cell">
                            <Box>
                              <Typography variant="body2">{formatDateTime(score.created_at)}</Typography>
                              <Typography variant="caption" color="text.secondary">
                                {formatRelativeTime(score.created_at, t)}
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
                    {t('userProfile.loadMore')} ({scoreHistory.length - displayedScoreCount} {t('userProfile.remaining')})
                  </Button>
                </Box>
              )}
                </>
              )}
            </Box>
          )}

          {/* Recent Changes Tab (Super Admin only) */}
          {activeTab === 1 && hasRole(Roles.SUPER_ADMIN) && (
            <Box sx={{ p: 2 }}>
              {loadingUserChanges ? (
                <Box sx={{ textAlign: 'center', py: 4 }}>
                  <Typography>{t('userProfile.loadingUserChanges')}</Typography>
                </Box>
              ) : userChanges.length === 0 ? (
                <Box sx={{ textAlign: 'center', py: 4 }}>
                  <Typography variant="body2" color="text.secondary">
                    {t('userProfile.noRecentChanges')}
                  </Typography>
                </Box>
              ) : (
                <>
                  <TableContainer>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell className="table-header-cell">{t('userProfile.changeType')}</TableCell>
                          <TableCell className="table-header-cell">{t('userProfile.member')}</TableCell>
                          <TableCell className="table-header-cell numeric-cell">{t('userProfile.date')}</TableCell>
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
                              <Chip
                                label={change.change_type}
                                size="small"
                                color={getChangeTypeColor(change.change_type)}
                              />
                            </TableCell>
                            <TableCell className="mixed-content-cell">
                              {change.member_name || (
                                <Typography variant="body2" color="text.secondary" className="numeric-cell">
                                  ID: {change.member_id} ({t('userProfile.deleted')})
                                </Typography>
                              )}
                            </TableCell>
                            <TableCell className="numeric-cell">
                              <Box>
                                <Typography variant="body2">{formatDateTime(change.changed_at)}</Typography>
                                <Typography variant="caption" color="text.secondary">
                                  {formatRelativeTime(change.changed_at, t)}
                                </Typography>
                              </Box>
                            </TableCell>
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
                        {t('userProfile.loadMore')} ({userChanges.length - displayedChangesCount} {t('userProfile.remaining')})
                      </Button>
                    </Box>
                  )}
                </>
              )}
            </Box>
          )}
        </Paper>
        </motion.div>

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
