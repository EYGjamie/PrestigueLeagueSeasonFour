-- Teams Tabelle
CREATE TABLE IF NOT EXISTS teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    division INTEGER NOT NULL,
    role_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index für schnellere Division-Abfragen
CREATE INDEX IF NOT EXISTS idx_teams_division ON teams(division);

-- Trigger für updated_at
CREATE TRIGGER IF NOT EXISTS update_teams_timestamp 
    AFTER UPDATE ON teams
BEGIN
    UPDATE teams SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Matches Tabelle
CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    division INTEGER NOT NULL,
    matchday INTEGER NOT NULL,
    team_home_id INTEGER NOT NULL,
    team_away_id INTEGER,
    score_home INTEGER CHECK(score_home IS NULL OR (score_home >= 0 AND score_home <= 4)),
    score_away INTEGER CHECK(score_away IS NULL OR (score_away >= 0 AND score_away <= 4)),
    channel_id TEXT,
    reported_at DATETIME,
    reported_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_home_id) REFERENCES teams(id),
    FOREIGN KEY (team_away_id) REFERENCES teams(id)
);

-- Indices für Matches
CREATE INDEX IF NOT EXISTS idx_matches_division ON matches(division);
CREATE INDEX IF NOT EXISTS idx_matches_matchday ON matches(matchday);
CREATE INDEX IF NOT EXISTS idx_matches_teams ON matches(team_home_id, team_away_id);
