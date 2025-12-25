import React, { useState, useCallback } from 'react';
import { Autocomplete, TextField, Box, Typography, CircularProgress, Avatar } from '@mui/material';
import { membersApi } from '../api';
import { MemberListItem } from '../types';
import { debounce, getGenderColor, getMemberPictureUrl } from '../utils/helpers';

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
  const [options, setOptions] = useState<MemberListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');

  // Debounced search function
  const searchMembers = useCallback(
    debounce(async (query: string) => {
      if (!query || query.length < 2) {
        setOptions([]);
        return;
      }

      setLoading(true);
      try {
        const result = await membersApi.searchMembers({
          name: query,
          limit: 20,
        });
        setOptions(result.members || []);
      } catch (error) {

        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300),
    []
  );

  const handleInputChange = (_: any, newInputValue: string) => {
    setInputValue(newInputValue);
    searchMembers(newInputValue);
  };

  return (
    <Autocomplete
      value={value}
      onChange={(_, newValue) => onChange(newValue)}
      inputValue={inputValue}
      onInputChange={handleInputChange}
      options={options || []}
      loading={loading}
      disabled={disabled}
      getOptionLabel={(option) => option.name}
      isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
      renderOption={(props, option) => (
        <Box component="li" {...props} sx={{ display: 'flex', gap: 1.5, alignItems: 'center' }}>
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
      noOptionsText={inputValue.length < 2 ? 'Type at least 2 characters' : 'No members found'}
    />
  );
};

export default MemberAutocomplete;
