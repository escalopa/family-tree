import React, { useEffect, useState } from 'react';
import { Box, CircularProgress, Container, Paper, Typography } from '@mui/material';
import { useParams } from 'react-router-dom';
import { familyTreesApi } from '../api';
import { PublicTreeResponse } from '../types';
import TreeVisualization from '../components/TreeVisualization';

const PublicTreePage: React.FC = () => {
  const { token = '' } = useParams<{ token: string }>();
  const [data, setData] = useState<PublicTreeResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        setData(await familyTreesApi.getPublicTree(token));
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [token]);

  if (loading) {
    return (
      <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Container maxWidth={false} sx={{ py: 4 }}>
      <Paper sx={{ p: 2, mb: 2 }}>
        <Typography variant="h4">Shared Family Tree</Typography>
        {data?.share && (
          <Typography variant="body2" color="text.secondary">
            Opened {data.share.visit_count}{data.share.max_visits ? ` / ${data.share.max_visits}` : ''} times
            {data.share.expires_at ? `, valid until ${data.share.expires_at}` : ''}
          </Typography>
        )}
      </Paper>
      {data?.tree ? (
        <TreeVisualization
          data={data.tree}
          onNodeClick={() => undefined}
          onSetRoot={() => undefined}
        />
      ) : (
        <Paper sx={{ p: 3 }}>
          <Typography>This shared tree is empty.</Typography>
        </Paper>
      )}
    </Container>
  );
};

export default PublicTreePage;
