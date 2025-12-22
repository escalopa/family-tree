import React, { useState } from 'react';
import {
  Card,
  CardContent,
  Avatar,
  Typography,
  Box,
  IconButton,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Grid,
} from '@mui/material';
import { Edit, Favorite, HeartBroken } from '@mui/icons-material';
import { SpouseInfo } from '../types';
import { formatDate } from '../utils/helpers';
import { spousesApi } from '../api';

interface SpouseCardProps {
  spouse: SpouseInfo;
  currentMemberId: number;
  onUpdate?: () => void;
  editable?: boolean;
}

const SpouseCard: React.FC<SpouseCardProps> = ({
  spouse,
  currentMemberId,
  onUpdate,
  editable = true,
}) => {
  const [openDialog, setOpenDialog] = useState(false);
  const [marriageDate, setMarriageDate] = useState(spouse.marriage_date || '');
  const [divorceDate, setDivorceDate] = useState(spouse.divorce_date || '');
  const [saving, setSaving] = useState(false);

  const handleOpenDialog = () => {
    setMarriageDate(spouse.marriage_date || '');
    setDivorceDate(spouse.divorce_date || '');
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await spousesApi.updateSpouseByMember({
        spouse_id: spouse.spouse_id,
        marriage_date: marriageDate || undefined,
        divorce_date: divorceDate || undefined,
      });
      handleCloseDialog();
      if (onUpdate) {
        onUpdate();
      }
    } catch (error) {
      console.error('Failed to update spouse:', error);
      alert('Failed to update spouse information');
    } finally {
      setSaving(false);
    }
  };

  const isDivorced = spouse.divorce_date !== null;

  return (
    <>
      <Card
        sx={{
          display: 'flex',
          alignItems: 'center',
          p: 2,
          mb: 2,
          border: isDivorced ? '1px solid #f44336' : '1px solid #4caf50',
          position: 'relative',
        }}
      >
        <Avatar
          src={spouse.picture || undefined}
          sx={{
            width: 60,
            height: 60,
            mr: 2,
            bgcolor: spouse.gender === 'M' ? '#00BCD4' : '#E91E63',
          }}
        >
          {spouse.english_name[0]}
        </Avatar>
        <CardContent sx={{ flex: 1, p: 0 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
            <Typography variant="h6">{spouse.english_name}</Typography>
            {isDivorced ? (
              <Chip
                icon={<HeartBroken />}
                label="Divorced"
                size="small"
                color="error"
              />
            ) : (
              <Chip
                icon={<Favorite />}
                label="Married"
                size="small"
                color="success"
              />
            )}
          </Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
            {spouse.arabic_name}
          </Typography>
          {spouse.marriage_date && (
            <Typography variant="caption" color="text.secondary">
              Married: {formatDate(spouse.marriage_date)}
            </Typography>
          )}
          {spouse.divorce_date && (
            <Typography variant="caption" color="text.secondary" display="block">
              Divorced: {formatDate(spouse.divorce_date)}
            </Typography>
          )}
        </CardContent>
        {editable && (
          <IconButton onClick={handleOpenDialog} color="primary">
            <Edit />
          </IconButton>
        )}
      </Card>

      {/* Edit Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Spouse Information</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                Editing marriage information for {spouse.english_name}
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Marriage Date"
                type="date"
                InputLabelProps={{ shrink: true }}
                value={marriageDate}
                onChange={(e) => setMarriageDate(e.target.value)}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Divorce Date"
                type="date"
                InputLabelProps={{ shrink: true }}
                value={divorceDate}
                onChange={(e) => setDivorceDate(e.target.value)}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} variant="contained" disabled={saving}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default SpouseCard;
