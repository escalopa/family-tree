
import React, { useState } from 'react';
import { enqueueSnackbar } from 'notistack';
import { useTranslation } from 'react-i18next';
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
  Box,
  Avatar,
} from '@mui/material';
import { MemberListItem } from '../types';
import { membersApi, spousesApi } from '../api';
import { getMemberPictureUrl, getGenderColor } from '../utils/helpers';

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
  const { t } = useTranslation();
  const [selectedSpouse, setSelectedSpouse] = useState<MemberListItem | null>(null);
  const [spouseOptions, setSpouseOptions] = useState<MemberListItem[]>([]);
  const [loadingSpouses, setLoadingSpouses] = useState(false);
  const [inputValue, setInputValue] = useState('');
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
      // Handle null or undefined members array from backend
      const members = result.members || [];
      // Filter out the current member
      const filteredMembers = members.filter(option => option.member_id !== memberId);
      console.log('Spouse search results:', filteredMembers); // Debug log
      setSpouseOptions(filteredMembers);
    } catch (error) {
      console.error('search for spouse:', error);
      setSpouseOptions([]);
    } finally {
      setLoadingSpouses(false);
    }
  };

  const handleSave = async () => {
    if (!selectedSpouse) {
      enqueueSnackbar(t('spouse.pleaseSelectSpouse'), { variant: 'warning' });
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
      const errorMessage = error?.response?.data?.error || t('spouse.failedToAddSpouse');
      enqueueSnackbar(errorMessage, { variant: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    setSelectedSpouse(null);
    setSpouseOptions([]);
    setInputValue('');
    setMarriageDate('');
    setDivorceDate('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('spouse.addSpouseFor', { name: memberName })}</DialogTitle>
      <DialogContent>
        <Grid container spacing={2} sx={{ mt: 1 }}>
          <Grid item xs={12}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              {oppositeGender === 'M' ? t('spouse.searchMaleSpouse') : t('spouse.searchFemaleSpouse')}
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <Autocomplete
              options={spouseOptions}
              getOptionLabel={(option) => option.name}
              loading={loadingSpouses}
              value={selectedSpouse}
              onChange={(_, newValue) => {
                console.log('Selected spouse:', newValue); // Debug log
                setSelectedSpouse(newValue);
              }}
              inputValue={inputValue}
              onInputChange={(_, newInputValue, reason) => {
                console.log('Input change:', newInputValue, 'reason:', reason); // Debug log
                setInputValue(newInputValue);
                if (reason === 'input') {
                  handleSearchSpouse(newInputValue);
                }
              }}
              filterOptions={(x) => x}
              isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('spouse.selectSpouse')}
                  placeholder={t('spouse.typeToSearch')}
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
                <Box component="li" {...props} key={option.member_id} sx={{ display: 'flex', gap: 1.5, alignItems: 'center' }}>
                  <Avatar
                    src={getMemberPictureUrl(option.member_id, option.picture) || undefined}
                    sx={{
                      width: 40,
                      height: 40,
                      bgcolor: getGenderColor(option.gender),
                    }}
                  >
                    {option.name.charAt(0) || '?'}
                  </Avatar>
                  <Box>
                    <Typography variant="body2">{option.name}</Typography>
                    {option.date_of_birth && (
                      <Typography variant="caption" color="text.secondary">
                        {option.date_of_birth}
                      </Typography>
                    )}
                  </Box>
                </Box>
              )}
              noOptionsText={inputValue.length < 2 ? t('spouse.typeToSearchMembers') : t('member.noMembers')}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label={t('spouse.marriageDate')}
              type="date"
              InputLabelProps={{ shrink: true }}
              value={marriageDate}
              onChange={(e) => setMarriageDate(e.target.value)}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label={t('spouse.divorceDate')}
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
          {t('common.cancel')}
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={saving || !selectedSpouse}>
          {saving ? t('spouse.adding') : t('spouse.addSpouse')}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default AddSpouseDialog;
