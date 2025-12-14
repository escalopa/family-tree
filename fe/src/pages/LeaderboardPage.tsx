import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Container, Typography, Card, CardContent, Avatar, Box, Chip } from '@mui/material';
import { Layout } from '../components/common/Layout';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { usersApi } from '../api/users';
import EmojiEventsIcon from '@mui/icons-material/EmojiEvents';

export const LeaderboardPage: React.FC = () => {
  const { data, isLoading } = useQuery({
    queryKey: ['leaderboard'],
    queryFn: () => usersApi.getLeaderboard(20),
  });

  if (isLoading) {
    return (
      <Layout>
        <LoadingSpinner />
      </Layout>
    );
  }

  const getRankColor = (rank: number) => {
    if (rank === 1) return 'gold';
    if (rank === 2) return 'silver';
    if (rank === 3) return '#cd7f32';
    return 'grey';
  };

  return (
    <Layout>
      <Container maxWidth="md">
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
          <EmojiEventsIcon sx={{ fontSize: 40, mr: 2, color: 'gold' }} />
          <Typography variant="h4">Leaderboard</Typography>
        </Box>
        <Typography variant="body1" color="textSecondary" paragraph>
          Top contributors to the family tree
        </Typography>

        {data?.users.map((user) => (
          <Card key={user.user_id} sx={{ mb: 2 }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Chip
                  label={`#${user.rank}`}
                  sx={{ bgcolor: getRankColor(user.rank), color: 'white', fontWeight: 'bold' }}
                />
                <Avatar src={user.avatar || undefined} alt={user.full_name} />
                <Box sx={{ flexGrow: 1 }}>
                  <Typography variant="h6">{user.full_name}</Typography>
                </Box>
                <Chip label={`${user.total_score} points`} color="primary" />
              </Box>
            </CardContent>
          </Card>
        ))}
      </Container>
    </Layout>
  );
};



