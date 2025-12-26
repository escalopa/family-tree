import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Avatar,
  Chip,
  CircularProgress,
} from '@mui/material';
import { EmojiEvents } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { usersApi } from '../api';
import { UserScore, Roles } from '../types';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';

const LeaderboardPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { hasRole } = useAuth();
  const { mode } = useTheme();
  const isDarkMode = mode === 'dark';
  const [leaderboard, setLeaderboard] = useState<UserScore[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadLeaderboard();
  }, []);

  const loadLeaderboard = async () => {
    try {
      const response = await usersApi.getLeaderboard();
      setLeaderboard(response.users);
    } catch (error) {

    } finally {
      setLoading(false);
    }
  };

  const getMedalColor = (rank: number) => {
    switch (rank) {
      case 1:
        return 'gold';
      case 2:
        return 'silver';
      case 3:
        return '#CD7F32'; // bronze
      default:
        return 'text.secondary';
    }
  };

  return (
    <Layout>
      <Box sx={{ width: '100%' }}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
          <Typography variant="h4" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
            <EmojiEvents sx={{ color: 'primary.main' }} />
            {t('leaderboard.title')}
          </Typography>
        </motion.div>

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 10 }}>
            <CircularProgress />
          </Box>
        ) : (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.1 }}
          >
          <TableContainer
            component={Paper}
            className={isDarkMode ? 'enhanced-table-dark' : 'enhanced-table'}
            sx={{
              borderRadius: '16px',
              boxShadow: '0 4px 20px rgba(0, 0, 0, 0.08)',
              overflow: 'hidden',
              ...(isDarkMode && {
                boxShadow: '0 4px 20px rgba(0, 0, 0, 0.3)',
              }),
              '& .MuiTableHead-root .MuiTableRow-root': {
                transition: 'none !important',
                '&:hover': {
                  backgroundColor: 'transparent !important',
                  transform: 'none !important',
                }
              }
            }}
          >
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell className="table-header-cell numeric-cell">{t('leaderboard.rank')}</TableCell>
                  <TableCell className="table-header-cell">{t('leaderboard.user')}</TableCell>
                  <TableCell className="table-header-cell numeric-cell">{t('leaderboard.totalScore')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {leaderboard.map((user) => {
                  const rankClass =
                    user.rank === 1 ? 'leaderboard-top-1' :
                    user.rank === 2 ? 'leaderboard-top-2' :
                    user.rank === 3 ? 'leaderboard-top-3' : '';

                  const medalClass =
                    user.rank === 1 ? 'rank-medal-gold' :
                    user.rank === 2 ? 'rank-medal-silver' :
                    user.rank === 3 ? 'rank-medal-bronze' : '';

                  return (
                    <TableRow
                      key={user.user_id}
                      hover={hasRole(Roles.ADMIN)}
                      className={rankClass}
                      sx={{
                        cursor: hasRole(Roles.ADMIN) ? 'pointer' : 'default',
                        transition: 'background-color 0.2s ease',
                      }}
                      onClick={() => hasRole(Roles.ADMIN) && navigate(`/users/${user.user_id}`)}
                    >
                      <TableCell className="numeric-cell">
                        <Box
                          sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1,
                            flexDirection: { xs: 'row', md: 'row' },
                            '[dir="rtl"] &': {
                              flexDirection: 'row-reverse'
                            }
                          }}
                        >
                          {user.rank <= 3 ? (
                            <EmojiEvents
                              className={medalClass}
                              sx={{ color: getMedalColor(user.rank), fontSize: 28 }}
                            />
                          ) : (
                            <Typography
                              variant="body1"
                              sx={{
                                fontWeight: 600,
                                color: 'text.secondary',
                                minWidth: '24px',
                                textAlign: 'center'
                              }}
                            >
                              {user.rank}
                            </Typography>
                          )}
                        </Box>
                      </TableCell>
                      <TableCell className="mixed-content-cell">
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                          <Avatar
                            src={user.avatar || undefined}
                            className={
                              user.rank === 1 ? 'avatar-ring-admin' :
                              isDarkMode ? 'enhanced-avatar-dark' : 'enhanced-avatar'
                            }
                            sx={{
                              width: user.rank <= 3 ? 50 : 40,
                              height: user.rank <= 3 ? 50 : 40,
                              transition: 'all 0.3s ease',
                              ...(user.rank === 1 && {
                                boxShadow: '0 0 0 4px rgba(255, 215, 0, 0.2), 0 4px 12px rgba(255, 215, 0, 0.4)',
                                border: '3px solid #FFD700'
                              }),
                              ...(user.rank === 2 && {
                                boxShadow: '0 0 0 4px rgba(192, 192, 192, 0.2), 0 4px 12px rgba(192, 192, 192, 0.4)',
                                border: '3px solid #C0C0C0'
                              }),
                              ...(user.rank === 3 && {
                                boxShadow: '0 0 0 4px rgba(205, 127, 50, 0.2), 0 4px 12px rgba(205, 127, 50, 0.4)',
                                border: '3px solid #CD7F32'
                              })
                            }}
                          >
                            {user.full_name[0]}
                          </Avatar>
                          <Box>
                            <Typography
                              variant="body1"
                              sx={{
                                fontWeight: user.rank <= 3 ? 600 : 400,
                                fontSize: user.rank === 1 ? '1.1rem' : '1rem'
                              }}
                            >
                              {user.full_name}
                            </Typography>
                            {user.rank <= 3 && (
                              <Typography
                                variant="caption"
                                sx={{
                                  color: user.rank === 1 ? '#FFD700' :
                                         user.rank === 2 ? '#C0C0C0' :
                                         '#CD7F32',
                                  fontWeight: 600,
                                  display: 'block'
                                }}
                              >
                                {user.rank === 1 ? '🏆 ' : user.rank === 2 ? '🥈 ' : '🥉 '}
                                {user.rank === 1 ? t('leaderboard.champion') :
                                 user.rank === 2 ? t('leaderboard.secondPlace') :
                                 t('leaderboard.thirdPlace')}
                              </Typography>
                            )}
                          </Box>
                        </Box>
                      </TableCell>
                      <TableCell className="numeric-cell">
                        <Chip
                          label={user.total_score}
                          color={user.rank <= 3 ? 'secondary' : 'primary'}
                          size={user.rank <= 3 ? 'medium' : 'small'}
                          variant={user.rank <= 3 ? 'filled' : 'outlined'}
                          className="enhanced-chip"
                          sx={{
                            fontWeight: user.rank <= 3 ? 700 : 500,
                            fontSize: user.rank === 1 ? '1rem' : '0.875rem',
                            ...(user.rank === 1 && {
                              background: 'linear-gradient(135deg, #FFD700 0%, #FFA500 100%)',
                              color: 'white',
                              boxShadow: '0 2px 8px rgba(255, 215, 0, 0.3)'
                            }),
                            ...(user.rank === 2 && {
                              background: 'linear-gradient(135deg, #C0C0C0 0%, #A8A8A8 100%)',
                              color: 'white',
                              boxShadow: '0 2px 8px rgba(192, 192, 192, 0.3)'
                            }),
                            ...(user.rank === 3 && {
                              background: 'linear-gradient(135deg, #CD7F32 0%, #B8733C 100%)',
                              color: 'white',
                              boxShadow: '0 2px 8px rgba(205, 127, 50, 0.3)'
                            })
                          }}
                        />
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
          </motion.div>
        )}
      </Box>
    </Layout>
  );
};

export default LeaderboardPage;
