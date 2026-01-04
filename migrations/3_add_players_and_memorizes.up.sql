-- 1. Update users table with missing email column
ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL DEFAULT '';

-- 2. Create status enum for memorizes
CREATE TYPE memorize_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

-- 3. Create players table
CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    teams_id INT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- 4. Create memorizes table
CREATE TABLE memorizes (
    id SERIAL PRIMARY KEY,
    matches_id INT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    players_id INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    surah INTEGER NOT NULL,
    status memorize_status NOT NULL DEFAULT 'PENDING'
);

-- 5. Add missing total_surah to tournaments
ALTER TABLE tournaments ADD COLUMN total_surah INTEGER NOT NULL DEFAULT 0;