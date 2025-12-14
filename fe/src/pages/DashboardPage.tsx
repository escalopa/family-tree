import React from 'react';
import { Container, Typography, Grid, Card, CardContent, Button } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { Layout } from '../components/common/Layout';
import { useAuthStore } from '../store/authStore';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import PeopleIcon from '@mui/icons-material/People';
import LeaderboardIcon from '@mui/icons-material/Leaderboard';

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();

  return (
    <Layout>
      <Container maxWidth="lg">
        <Typography variant="h4" gutterBottom>
          Welcome, {user?.full_name}!
        </Typography>
        <Typography variant="body1" color="textSecondary" paragraph>
          Manage your family tree and explore your ancestry.
        </Typography>

        <Grid container spacing={3} sx={{ mt: 2 }}>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent sx={{ textAlign: 'center', py: 4 }}>
                <AccountTreeIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
                <Typography variant="h6" gutterBottom>
                  View Family Tree
                </Typography>
                <Typography variant="body2" color="textSecondary" paragraph>
                  Explore your family tree in an interactive diagram
                </Typography>
                <Button variant="contained" onClick={() => navigate('/tree')}>
                  Go to Tree
                </Button>
              </CardContent>
            </Card>
          </Grid>

          {user && user.role_id >= 300 && (
            <Grid item xs={12} md={4}>
              <Card>
                <CardContent sx={{ textAlign: 'center', py: 4 }}>
                  <PeopleIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
                  <Typography variant="h6" gutterBottom>
                    Manage Members
                  </Typography>
                  <Typography variant="body2" color="textSecondary" paragraph>
                    Add, edit, and manage family members
                  </Typography>
                  <Button variant="contained" onClick={() => navigate('/members')}>
                    Manage Members
                  </Button>
                </CardContent>
              </Card>
            </Grid>
          )}

          <Grid item xs={12} md={4}>
            <Card>
              <CardContent sx={{ textAlign: 'center', py: 4 }}>
                <LeaderboardIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
                <Typography variant="h6" gutterBottom>
                  Leaderboard
                </Typography>
                <Typography variant="body2" color="textSecondary" paragraph>
                  See who's contributing the most to the family tree
                </Typography>
                <Button variant="contained" onClick={() => navigate('/leaderboard')}>
                  View Leaderboard
                </Button>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Container>
    </Layout>
  );
};



