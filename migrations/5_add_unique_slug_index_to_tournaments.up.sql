CREATE UNIQUE INDEX idx_tournaments_slug ON tournaments(slug) WHERE deleted_at IS NULL;
