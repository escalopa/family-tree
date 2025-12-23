import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Autocomplete,
  TextField,
  Alert,
  CircularProgress,
} from '@mui/material';
import { Search } from '@mui/icons-material';
import { Member } from '../types';

interface RelationFinderProps {
  members: Member[];
  onFindRelation: (member1Id: number, member2Id: number) => Promise<void>;
  loading?: boolean;
}

const RelationFinder: React.FC<RelationFinderProps> = ({ members, onFindRelation, loading }) => {
  const [member1, setMember1] = useState<Member | null>(null);
  const [member2, setMember2] = useState<Member | null>(null);
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
        family tree.
      </Typography>

      <Box sx={{ display: 'flex', gap: 2, alignItems: 'flex-start', flexWrap: 'wrap' }}>
        <Box sx={{ flex: '1 1 300px', minWidth: 250 }}>
          <Autocomplete
            options={members}
            value={member1}
            onChange={(_, newValue) => {
              setMember1(newValue);
              setError(null);
            }}
            getOptionLabel={(option) => `${option.arabic_name} (${option.english_name})`}
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box>
                  <Typography variant="body2">{option.arabic_name}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {option.english_name}
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => <TextField {...params} label="First Member" />}
            isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
          />
        </Box>

        <Box sx={{ flex: '1 1 300px', minWidth: 250 }}>
          <Autocomplete
            options={members}
            value={member2}
            onChange={(_, newValue) => {
              setMember2(newValue);
              setError(null);
            }}
            getOptionLabel={(option) => `${option.arabic_name} (${option.english_name})`}
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box>
                  <Typography variant="body2">{option.arabic_name}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {option.english_name}
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => <TextField {...params} label="Second Member" />}
            isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
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

