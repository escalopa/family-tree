-- Revert user_language_preferences table to use primary and secondary languages
ALTER TABLE user_language_preferences
    DROP CONSTRAINT IF EXISTS fk_user_lang_pref_preferred,
    DROP COLUMN IF EXISTS preferred_language,
    ADD COLUMN primary_language VARCHAR(10) NOT NULL DEFAULT 'ar',
    ADD COLUMN secondary_language VARCHAR(10) NOT NULL DEFAULT 'en';

-- Add foreign key constraints
ALTER TABLE user_language_preferences
    ADD CONSTRAINT fk_user_lang_pref_primary FOREIGN KEY (primary_language) REFERENCES languages(language_code),
    ADD CONSTRAINT fk_user_lang_pref_secondary FOREIGN KEY (secondary_language) REFERENCES languages(language_code),
    ADD CONSTRAINT chk_user_lang_pref_different CHECK (primary_language != secondary_language);

-- Drop the preferred_language index
DROP INDEX IF EXISTS idx_user_language_preferences_preferred_language;
