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
import { Edit, Favorite, HeartBroken, Delete } from '@mui/icons-material';
import { SpouseInfo } from '../types';
import { formatDate, getMemberPictureUrl } from '../utils/helpers';
import { spousesApi } from '../api';

interface SpouseCardProps {
  spouse: SpouseInfo;
  currentMemberId: number;
  onUpdate?: () => void;
  editable?: boolean;
  onMemberClick?: () => void;
}

const SpouseCard: React.FC<SpouseCardProps> = ({
  spouse,
  currentMemberId,
  onUpdate,
  editable = true,
  onMemberClick,
}) => {
  const [openDialog, setOpenDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const [marriageDate, setMarriageDate] = useState(spouse.marriage_date || '');
  const [divorceDate, setDivorceDate] = useState(spouse.divorce_date || '');
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

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

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await spousesApi.removeSpouse(spouse.spouse_id);
      setOpenDeleteDialog(false);
      if (onUpdate) {
        onUpdate();
      }
    } catch (error: any) {
      console.error('Failed to delete spouse:', error);
      const errorMessage = error?.response?.data?.error || 'Failed to delete spouse relationship';
      alert(errorMessage);
    } finally {
      setDeleting(false);
    }
  };

  const isDivorced = spouse.divorce_date !== null;

  const handleCardClick = (e: React.MouseEvent) => {
    // Don't trigger if clicking on edit/delete buttons
    const target = e.target as HTMLElement;
    if (target.closest('button')) {
      return;
    }
    onMemberClick?.();
  };

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
          cursor: onMemberClick ? 'pointer' : 'default',
          '&:hover': onMemberClick ? { boxShadow: 3, bgcolor: 'action.hover' } : {},
        }}
        onClick={handleCardClick}
      >
        <Avatar
          src={getMemberPictureUrl(spouse.member_id, spouse.picture) || undefined}
          sx={{
            width: 60,
            height: 60,
            mr: 2,
            bgcolor: spouse.gender === 'M' ? '#00BCD4' : '#E91E63',
          }}
        >
          {spouse.name?.[0] || '?'}
        </Avatar>
        <CardContent sx={{ flex: 1, p: 0 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
            <Typography variant="h6">{spouse.name || 'N/A'}</Typography>
            {isDivorced && (
              <Chip
                icon={<HeartBroken />}
                label="Divorced"
                size="small"
                color="error"
              />
            )}
          </Box>
          {spouse.marriage_date && (
            <Typography variant="caption" color="text.secondary">
              Married: {formatDate(spouse.marriage_date)}
              {spouse.married_years !== null && spouse.married_years !== undefined && (
                <> ({spouse.married_years} {spouse.married_years === 1 ? 'year' : 'years'})</>
              )}
            </Typography>
          )}
          {spouse.divorce_date && (
            <Typography variant="caption" color="text.secondary" display="block">
              Divorced: {formatDate(spouse.divorce_date)}
            </Typography>
          )}
        </CardContent>
        {editable && (
          <Box>
            <IconButton onClick={handleOpenDialog} color="primary">
              <Edit />
            </IconButton>
            <IconButton onClick={() => setOpenDeleteDialog(true)} color="error">
              <Delete />
            </IconButton>
          </Box>
        )}
      </Card>

      {/* Edit Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Spouse Information</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                Editing marriage information for {spouse.name || 'Unknown'}
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

      {/* Delete Confirmation Dialog */}
      <Dialog open={openDeleteDialog} onClose={() => setOpenDeleteDialog(false)}>
        <DialogTitle>Delete Spouse Relationship</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete the spouse relationship with {spouse.name || 'Unknown'}?
          </Typography>
          <Typography variant="body2" color="error" sx={{ mt: 2 }}>
            Note: You cannot delete a spouse relationship if there are children with both parents.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDeleteDialog(false)} disabled={deleting}>
            Cancel
          </Button>
          <Button onClick={handleDelete} variant="contained" color="error" disabled={deleting}>
            {deleting ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default SpouseCard;
