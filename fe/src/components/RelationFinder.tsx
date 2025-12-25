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
import { useTranslation } from 'react-i18next';
import { MemberListItem } from '../types';
import MemberAutocomplete from './MemberAutocomplete';

interface RelationFinderProps {
  onFindRelation: (member1Id: number, member2Id: number) => Promise<void>;
  loading?: boolean;
}

const RelationFinder: React.FC<RelationFinderProps> = ({ onFindRelation, loading }) => {
  const { t, i18n } = useTranslation();
  const isRTL = i18n.dir() === 'rtl';
  const [member1, setMember1] = useState<MemberListItem | null>(null);
  const [member2, setMember2] = useState<MemberListItem | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleFindRelation = async () => {
    if (!member1 || !member2) {
      setError(t('tree.selectBothMembers'));
      return;
    }

    if (member1.member_id === member2.member_id) {
      setError(t('tree.selectTwoDifferentMembers'));
      return;
    }

    setError(null);
    try {
      await onFindRelation(member1.member_id, member2.member_id);
    } catch (err) {
      setError(t('tree.failedToFindRelation'));
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        {t('tree.findRelationBetweenMembers')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        {t('tree.findRelationDescription')}
      </Typography>

      <Box sx={{ display: 'flex', gap: 2, alignItems: 'flex-start', flexWrap: 'wrap' }}>
        <Box sx={{ flex: '1 1 300px', minWidth: 250 }}>
          <MemberAutocomplete
            label={t('tree.firstMember')}
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
            label={t('tree.secondMember')}
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
          {t('tree.findRelation')}
        </Button>
      </Box>

      {error && (
        <Alert
          severity="error"
          sx={{
            mt: 2,
            textAlign: isRTL ? 'right' : 'left',
            '& .MuiAlert-icon': {
              marginInlineEnd: 1.5,
              marginInlineStart: 0,
            }
          }}
        >
          {error}
        </Alert>
      )}
    </Paper>
  );
};

export default RelationFinder;
