import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Container,
  Typography,
  Box,
  ToggleButtonGroup,
  ToggleButton,
  TextField,
  Paper,
} from '@mui/material';
import { Layout } from '../components/common/Layout';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { treeApi } from '../api/tree';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import ViewListIcon from '@mui/icons-material/ViewList';

export const TreePage: React.FC = () => {
  const [style, setStyle] = useState<'tree' | 'list'>('tree');
  const [rootId, setRootId] = useState<number | undefined>();

  const { data, isLoading } = useQuery({
    queryKey: ['tree', style, rootId],
    queryFn: () => treeApi.getTree({ style, root: rootId }),
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
      <Container maxWidth="xl">
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4">Family Tree</Typography>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <TextField
              label="Root Member ID"
              type="number"
              size="small"
              value={rootId || ''}
              onChange={(e) => setRootId(e.target.value ? Number(e.target.value) : undefined)}
            />
            <ToggleButtonGroup
              value={style}
              exclusive
              onChange={(_, newStyle) => newStyle && setStyle(newStyle)}
            >
              <ToggleButton value="tree">
                <AccountTreeIcon sx={{ mr: 1 }} /> Tree
              </ToggleButton>
              <ToggleButton value="list">
                <ViewListIcon sx={{ mr: 1 }} /> List
              </ToggleButton>
            </ToggleButtonGroup>
          </Box>
        </Box>

        <Paper sx={{ p: 3, minHeight: '500px' }}>
          <Typography variant="body2" color="textSecondary">
            Tree visualization will be implemented with ReactFlow
          </Typography>
          <pre>{JSON.stringify(data, null, 2)}</pre>
        </Paper>
      </Container>
    </Layout>
  );
};



