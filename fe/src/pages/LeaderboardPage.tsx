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
} from '@mui/material';
import { EmojiEvents } from '@mui/icons-material';
import { usersApi } from '../api';
import { UserScore, Roles } from '../types';
import Layout from '../components/Layout/Layout';
import { useAuth } from '../contexts/AuthContext';

const LeaderboardPage: React.FC = () => {
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
      console.error('Failed to load leaderboard:', error);
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
      <Box>
        <Typography variant="h4" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <EmojiEvents />
          Leaderboard
        </Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          Top contributors to the family tree
        </Typography>

        {loading ? (
          <Typography>Loading...</Typography>
        ) : (
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Rank</TableCell>
                  <TableCell>User</TableCell>
                  <TableCell align="right">Total Score</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {leaderboard.map((user) => (
                  <TableRow
                    key={user.user_id}
                    hover={hasRole(Roles.ADMIN)}
                    sx={{ cursor: hasRole(Roles.ADMIN) ? 'pointer' : 'default' }}
                    onClick={() => hasRole(Roles.ADMIN) && navigate(`/users/${user.user_id}`)}
                  >
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {user.rank <= 3 ? (
                          <EmojiEvents sx={{ color: getMedalColor(user.rank) }} />
                        ) : (
                          <Typography variant="body1">{user.rank}</Typography>
                        )}
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Avatar src={user.avatar || undefined}>
                          {user.full_name[0]}
                        </Avatar>
                        <Box>
                          <Typography variant="body1">{user.full_name}</Typography>
                        </Box>
                      </Box>
                    </TableCell>
                    <TableCell align="right">
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
        )}
      </Box>
    </Layout>
  );
};

export default LeaderboardPage;
