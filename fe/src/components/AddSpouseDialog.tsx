import React, { useState } from 'react';
import { enqueueSnackbar } from 'notistack';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Grid,
  Typography,
  Autocomplete,
  CircularProgress,
} from '@mui/material';
import { MemberListItem } from '../types';
import { membersApi, spousesApi } from '../api';

interface AddSpouseDialogProps {
  open: boolean;
  onClose: () => void;
  memberId: number;
  memberName: string;
  memberGender: 'M' | 'F';
  onSuccess: () => void;
}

const AddSpouseDialog: React.FC<AddSpouseDialogProps> = ({
  open,
  onClose,
  memberId,
  memberName,
  memberGender,
  onSuccess,
}) => {
  const [selectedSpouse, setSelectedSpouse] = useState<MemberListItem | null>(null);
  const [spouseOptions, setSpouseOptions] = useState<MemberListItem[]>([]);
  const [loadingSpouses, setLoadingSpouses] = useState(false);
  const [marriageDate, setMarriageDate] = useState('');
  const [divorceDate, setDivorceDate] = useState('');
  const [saving, setSaving] = useState(false);

  const oppositeGender = memberGender === 'M' ? 'F' : 'M';

  const handleSearchSpouse = async (query: string) => {
    if (query.length < 2) {
      setSpouseOptions([]);
      return;
    }

    setLoadingSpouses(true);
    try {
      const result = await membersApi.searchMembers({
        name: query,
        gender: oppositeGender,
        limit: 20,
      });
      // Filter out the current member
      setSpouseOptions(result.members.filter(option => option.member_id !== memberId));
    } catch (error) {
      console.error('search for spouse:', error);
    } finally {
      setLoadingSpouses(false);
    }
  };

  const handleSave = async () => {
    if (!selectedSpouse) {
      enqueueSnackbar('Please select a spouse', { variant: 'warning' });
      return;
    }

    setSaving(true);
    try {
      // Determine father_id and mother_id based on gender
      const fatherId = memberGender === 'M' ? memberId : selectedSpouse.member_id;
      const motherId = memberGender === 'F' ? memberId : selectedSpouse.member_id;

      await spousesApi.addSpouse({
        father_id: fatherId,
        mother_id: motherId,
        marriage_date: marriageDate || undefined,
        divorce_date: divorceDate || undefined,
      });
      onSuccess();
      handleClose();
    } catch (error: any) {
      console.error('add spouse:', error);
      const errorMessage = error?.response?.data?.error || 'Failed to add spouse relationship. They may already be connected.';
      enqueueSnackbar(errorMessage, { variant: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    setSelectedSpouse(null);
    setSpouseOptions([]);
    setMarriageDate('');
    setDivorceDate('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Add Spouse for {memberName}</DialogTitle>
      <DialogContent>
        <Grid container spacing={2} sx={{ mt: 1 }}>
          <Grid item xs={12}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Search for a {oppositeGender === 'M' ? 'male' : 'female'} member to add as spouse
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <Autocomplete
              options={spouseOptions}
              getOptionLabel={(option) => option.name}
              loading={loadingSpouses}
              value={selectedSpouse}
              onChange={(_, newValue) => setSelectedSpouse(newValue)}
              onInputChange={(_, newInputValue) => {
                handleSearchSpouse(newInputValue);
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Select Spouse"
                  placeholder="Type to search..."
                  required
                  InputProps={{
                    ...params.InputProps,
                    endAdornment: (
                      <>
                        {loadingSpouses ? <CircularProgress size={20} /> : null}
                        {params.InputProps.endAdornment}
                      </>
                    ),
                  }}
                />
              )}
              renderOption={(props, option) => (
                <li {...props} key={option.member_id}>
                  <div>
                    <div>{option.name}</div>
                  </div>
                </li>
              )}
              noOptionsText="Type to search for members"
            />
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
        <Button onClick={handleClose} disabled={saving}>
          Cancel
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={saving || !selectedSpouse}>
          {saving ? 'Adding...' : 'Add Spouse'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default AddSpouseDialog;
