-- Drop tables first (child tables first)
DROP TABLE IF EXISTS memorizes;
DROP TABLE IF EXISTS players;

-- Drop custom types
DROP TYPE IF EXISTS memorize_status;

-- Remove columns from existing tables
ALTER TABLE tournaments DROP COLUMN IF EXISTS total_surah;
ALTER TABLE users DROP COLUMN IF EXISTS email;