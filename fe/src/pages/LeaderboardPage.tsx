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

const LeaderboardPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { hasRole } = useAuth();
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
          initial={{ opacity: 0, y: -20 }}
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
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell className="table-header-cell numeric-cell">{t('leaderboard.rank')}</TableCell>
                    <TableCell className="table-header-cell">{t('leaderboard.user')}</TableCell>
                    <TableCell className="table-header-cell numeric-cell">{t('leaderboard.totalScore')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {leaderboard.map((user, index) => (
                    <TableRow
                      key={user.user_id}
                      hover={hasRole(Roles.ADMIN)}
                      component={motion.tr}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ duration: 0.3, delay: index * 0.05 }}
                      sx={{
                        cursor: hasRole(Roles.ADMIN) ? 'pointer' : 'default',
                        '&:hover': {
                          bgcolor: 'action.hover',
                          transform: 'translateX(4px)',
                          transition: 'all 0.2s ease-in-out',
                        },
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
                            <motion.div
                              initial={{ scale: 0 }}
                              animate={{ scale: 1 }}
                              transition={{ duration: 0.3, delay: index * 0.05 + 0.2 }}
                            >
                              <EmojiEvents sx={{ color: getMedalColor(user.rank) }} />
                            </motion.div>
                          ) : (
                            <Typography variant="body1">{user.rank}</Typography>
                          )}
                        </Box>
                      </TableCell>
                      <TableCell className="mixed-content-cell">
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                          <Avatar src={user.avatar || undefined}>
                            {user.full_name[0]}
                          </Avatar>
                          <Box>
                            <Typography variant="body1">{user.full_name}</Typography>
                          </Box>
                        </Box>
                      </TableCell>
                      <TableCell className="numeric-cell">
                        <Chip
                          label={user.total_score}
                          color="primary"
                          size="small"
                        />
                      </TableCell>
                    </TableRow>
                  ))}
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
