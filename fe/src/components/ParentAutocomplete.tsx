import React, { useState, useEffect } from 'react';
import { Autocomplete, TextField, Avatar, Box, Typography, CircularProgress } from '@mui/material';
import { membersApi } from '../api';
import { getMemberPictureUrl } from '../utils/helpers';
import { ParentOption } from '../types';

interface ParentAutocompleteProps {
  label: string;
  gender: 'M' | 'F';
  value: number | null;
  onChange: (value: number | null) => void;
  disabled?: boolean;
}

const ParentAutocomplete: React.FC<ParentAutocompleteProps> = ({
  label,
  gender,
  value,
  onChange,
  disabled = false,
}) => {
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
        const results = await membersApi.searchParents(inputValue, gender);
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
      // Try to find in current options first
      const found = options.find(opt => opt.member_id === value);
      if (found) {
        setSelectedOption(found);
      }
    }
  }, [value, options, selectedOption]);

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
      getOptionLabel={(option) => `${option.arabic_name} - ${option.english_name}`}
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
        <Box component="li" {...props} key={option.member_id}>
          <Avatar
            src={getMemberPictureUrl(option.member_id, option.picture) || undefined}
            sx={{
              width: 32,
              height: 32,
              mr: 2,
              bgcolor: option.gender === 'M' ? '#00BCD4' : '#E91E63',
            }}
          >
            {option.english_name[0]}
          </Avatar>
          <Box>
            <Typography variant="body2">{option.arabic_name}</Typography>
            <Typography variant="caption" color="text.secondary">
              {option.english_name}
            </Typography>
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
