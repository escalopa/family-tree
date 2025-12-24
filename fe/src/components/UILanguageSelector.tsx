import React from 'react';
import { Select, MenuItem, FormControl, InputLabel, SelectChangeEvent } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useUILanguage } from '../contexts/UILanguageContext';

const UILanguageSelector: React.FC = () => {
  const { t } = useTranslation();
  const { uiLanguage, supportedLanguages, changeUILanguage } = useUILanguage();

  const handleChange = (event: SelectChangeEvent<string>) => {
    changeUILanguage(event.target.value);
  };

  const languageNames: Record<string, string> = {
    en: 'English',
    ar: 'العربية',
    ru: 'Русский',
  };

  return (
    <FormControl size="small" sx={{ minWidth: 120 }}>
      <InputLabel id="ui-language-select-label">{t('language.uiLanguage')}</InputLabel>
      <Select
        labelId="ui-language-select-label"
        id="ui-language-select"
        value={uiLanguage}
        label={t('language.uiLanguage')}
        onChange={handleChange}
      >
        {supportedLanguages.map((lang) => (
          <MenuItem key={lang} value={lang}>
            {languageNames[lang] || lang}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};

export default UILanguageSelector;
