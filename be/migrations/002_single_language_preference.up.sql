-- Update user_language_preferences table to use single preferred language
ALTER TABLE user_language_preferences
    DROP CONSTRAINT IF EXISTS chk_user_lang_pref_different,
    DROP CONSTRAINT IF EXISTS fk_user_lang_pref_primary,
    DROP CONSTRAINT IF EXISTS fk_user_lang_pref_secondary,
    DROP COLUMN IF EXISTS primary_language,
    DROP COLUMN IF EXISTS secondary_language,
    ADD COLUMN preferred_language VARCHAR(10) NOT NULL DEFAULT 'en';

-- Add foreign key constraint for preferred_language
ALTER TABLE user_language_preferences
    ADD CONSTRAINT fk_user_lang_pref_preferred FOREIGN KEY (preferred_language) REFERENCES languages(language_code);

-- Create index on preferred_language
CREATE INDEX IF NOT EXISTS idx_user_language_preferences_preferred_language ON user_language_preferences(preferred_language);
