import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Alert,
  CircularProgress,
} from '@mui/material';
import { Search } from '@mui/icons-material';
import { MemberListItem } from '../types';
import MemberAutocomplete from './MemberAutocomplete';

interface RelationFinderProps {
  onFindRelation: (member1Id: number, member2Id: number) => Promise<void>;
  loading?: boolean;
}

const RelationFinder: React.FC<RelationFinderProps> = ({ onFindRelation, loading }) => {
  const [member1, setMember1] = useState<MemberListItem | null>(null);
  const [member2, setMember2] = useState<MemberListItem | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleFindRelation = async () => {
    if (!member1 || !member2) {
      setError('Please select both members');
      return;
    }

    if (member1.member_id === member2.member_id) {
      setError('Please select two different members');
      return;
    }

    setError(null);
    try {
      await onFindRelation(member1.member_id, member2.member_id);
    } catch (err) {
      setError('Failed to find relation between members');
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        Find Relation Between Two Members
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Select two family members to see their relationship and the path connecting them in the
        family tree. Start typing to search.
      </Typography>

      <Box sx={{ display: 'flex', gap: 2, alignItems: 'flex-start', flexWrap: 'wrap' }}>
        <Box sx={{ flex: '1 1 300px', minWidth: 250 }}>
          <MemberAutocomplete
            label="First Member"
            value={member1}
            onChange={(value) => {
              setMember1(value);
              setError(null);
            }}
            disabled={loading}
          />
        </Box>

        <Box sx={{ flex: '1 1 300px', minWidth: 250 }}>
          <MemberAutocomplete
            label="Second Member"
            value={member2}
            onChange={(value) => {
              setMember2(value);
              setError(null);
            }}
            disabled={loading}
          />
        </Box>

        <Button
          variant="contained"
          startIcon={loading ? <CircularProgress size={20} /> : <Search />}
          onClick={handleFindRelation}
          disabled={loading || !member1 || !member2}
          sx={{ minWidth: 150, height: 56 }}
        >
          Find Relation
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error}
        </Alert>
      )}
    </Paper>
  );
};

export default RelationFinder;
