import React, { useState, useEffect } from 'react';
import { Autocomplete, TextField, Avatar, Box, Typography, CircularProgress } from '@mui/material';
import { membersApi } from '../api';
import { getMemberPictureUrl } from '../utils/helpers';
import { ParentOption } from '../types';
import { useLanguage } from '../contexts/LanguageContext';

interface ParentAutocompleteProps {
  label: string;
  gender: 'M' | 'F';
  value: number | null;
  onChange: (value: number | null) => void;
  disabled?: boolean;
  initialParent?: { member_id: number; names: Record<string, string>; picture: string | null } | null;
}

const ParentAutocomplete: React.FC<ParentAutocompleteProps> = ({
  label,
  gender,
  value,
  onChange,
  disabled = false,
  initialParent = null,
}) => {
  const { getPreferredName, getAllNamesFormatted } = useLanguage();
  const [options, setOptions] = useState<ParentOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [selectedOption, setSelectedOption] = useState<ParentOption | null>(null);

  // Debounced search
  useEffect(() => {
    if (inputValue.length < 2) {
      setOptions([]);
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const results = await membersApi.searchMemberInfo(inputValue, gender);
        setOptions(results || []);
      } catch (error) {
        console.error('Failed to search parents:', error);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [inputValue, gender]);

  // Load initial selected option
  useEffect(() => {
    if (value && !selectedOption) {
      // First check if we have an initial parent to display
      if (initialParent && initialParent.member_id === value) {
        setSelectedOption({ ...initialParent, gender } as ParentOption);
      } else {
        // Try to find in current options
        const found = options.find(opt => opt.member_id === value);
        if (found) {
          setSelectedOption(found);
        }
      }
    } else if (!value && selectedOption) {
      // Clear selection if value becomes null/undefined
      setSelectedOption(null);
    }
  }, [value, options, selectedOption, initialParent, gender]);

  return (
    <Autocomplete
      options={options || []}
      value={selectedOption}
      onChange={(_, newValue) => {
        setSelectedOption(newValue);
        onChange(newValue ? newValue.member_id : null);
      }}
      inputValue={inputValue}
      onInputChange={(_, newInputValue, reason) => {
        // Clear selectedOption when user clears the input
        if (reason === 'clear') {
          setSelectedOption(null);
          onChange(null);
        }
        setInputValue(newInputValue);
      }}
      getOptionLabel={(option) => getPreferredName(option)}
      isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
      loading={loading}
      disabled={disabled}
      freeSolo={false}
      clearOnEscape
      clearOnBlur={false}
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder="Type to search..."
          InputProps={{
            ...params.InputProps,
            startAdornment: selectedOption && (
              <Avatar
                src={getMemberPictureUrl(selectedOption.member_id, selectedOption.picture) || undefined}
                sx={{
                  width: 32,
                  height: 32,
                  bgcolor: gender === 'M' ? '#00BCD4' : '#E91E63',
                  mr: 1,
                }}
              >
                {getPreferredName(selectedOption).charAt(0) || '?'}
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
      renderOption={(props, option) => (
        <Box component="li" {...props} key={option.member_id} sx={{ display: 'flex', gap: 1.5, alignItems: 'center' }}>
          <Avatar
            src={getMemberPictureUrl(option.member_id, option.picture) || undefined}
            sx={{
              width: 40,
              height: 40,
              bgcolor: option.gender === 'M' ? '#00BCD4' : '#E91E63',
            }}
          >
            {getPreferredName(option).charAt(0) || '?'}
          </Avatar>
          <Box>
            <Typography variant="body2">{getAllNamesFormatted(option)}</Typography>
          </Box>
        </Box>
      )}
      noOptionsText={
        inputValue.length < 2
          ? 'Type at least 2 characters to search'
          : 'No results found'
      }
    />
  );
};

export default ParentAutocomplete;
