import React, { useState, useEffect } from 'react';
import { Autocomplete, TextField, Box, Typography, CircularProgress, Avatar } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { membersApi } from '../api';
import { MemberListItem } from '../types';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';

interface MemberAutocompleteProps {
  label: string;
  value: MemberListItem | null;
  onChange: (value: MemberListItem | null) => void;
  disabled?: boolean;
}

const MemberAutocomplete: React.FC<MemberAutocompleteProps> = ({
  label,
  value,
  onChange,
  disabled,
}) => {
  const { t } = useTranslation();
  const [options, setOptions] = useState<MemberListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');

  // Debounced search with useEffect
  useEffect(() => {
    if (inputValue.length < 2) {
      // Don't clear options immediately - let user see previous results
      if (inputValue.length === 0) {
        setOptions([]);
      }
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const result = await membersApi.searchMembers({
          name: inputValue,
          limit: 20,
        });
        console.log('Search result:', result);
        if (Array.isArray(result.members)) {
          setOptions(result.members);
        } else {
          console.warn('Search results is not an array:', result);
          setOptions([]);
        }
      } catch (error) {
        console.error('Failed to search members:', error);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [inputValue]);

  return (
    <Autocomplete
      options={options}
      value={value}
      onChange={(_, newValue) => {
        onChange(newValue);
      }}
      inputValue={inputValue}
      onInputChange={(_, newInputValue, reason) => {
        // Clear value when user clears the input
        if (reason === 'clear') {
          onChange(null);
          setInputValue('');
        } else {
          setInputValue(newInputValue);
        }
      }}
      getOptionLabel={(option) => option.name}
      isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
      filterOptions={(x) => x} // Don't filter options - backend already filtered them
      loading={loading}
      disabled={disabled}
      freeSolo={false}
      clearOnEscape
      clearOnBlur={false}
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
          </Box>
        </Box>
      )}
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder={value ? '' : t('member.searchByName')}
          helperText={value ? `${t('language.selected')}: ${value.name}` : undefined}
          InputProps={{
            ...params.InputProps,
            startAdornment: value && (
              <Avatar
                src={getMemberPictureUrl(value.member_id, value.picture) || undefined}
                sx={{
                  width: 32,
                  height: 32,
                  bgcolor: getGenderColor(value.gender),
                  mr: 1,
                }}
              >
                {value.name.charAt(0) || '?'}
              </Avatar>
            ),
            endAdornment: (
              <>
                {loading ? <CircularProgress color="inherit" size={20} /> : null}
                {params.InputProps.endAdornment}
              </>
            ),
          }}
        />
      )}
      noOptionsText={
        loading
          ? t('common.loading')
          : inputValue.length < 2
          ? t('member.typeAtLeastTwoCharacters')
          : t('member.noMembersFound')
      }
    />
  );
};

export default MemberAutocomplete;
