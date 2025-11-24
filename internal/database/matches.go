package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Match repräsentiert ein Spiel in der Datenbank
type Match struct {
	ID         int
	Division   int
	Matchday   int
	TeamHomeID int
	TeamAwayID sql.NullInt64
	ScoreHome  sql.NullInt64
	ScoreAway  sql.NullInt64
	ChannelID  sql.NullString
	ReportedAt sql.NullTime
	ReportedBy sql.NullString
	CreatedAt  time.Time
}

// CreateMatch erstellt ein neues Match
func (d *Database) CreateMatch(division, matchday, teamHomeID int, teamAwayID *int) (*Match, error) {
	var awayID sql.NullInt64
	if teamAwayID != nil {
		awayID = sql.NullInt64{Int64: int64(*teamAwayID), Valid: true}
	}

	result, err := d.DB.Exec(
		"INSERT INTO matches (division, matchday, team_home_id, team_away_id) VALUES (?, ?, ?, ?)",
		division, matchday, teamHomeID, awayID,
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Matches: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Match-ID: %w", err)
	}

	return d.GetMatchByID(int(id))
}

// GetMatchByID ruft ein Match anhand der ID ab
func (d *Database) GetMatchByID(id int) (*Match, error) {
	match := &Match{}
	err := d.DB.QueryRow(
		`SELECT id, division, matchday, team_home_id, team_away_id, 
		 score_home, score_away, channel_id, reported_at, reported_by, created_at 
		 FROM matches WHERE id = ?`,
		id,
	).Scan(
		&match.ID, &match.Division, &match.Matchday, &match.TeamHomeID, &match.TeamAwayID,
		&match.ScoreHome, &match.ScoreAway, &match.ChannelID, &match.ReportedAt,
		&match.ReportedBy, &match.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("match mit ID %d nicht gefunden", id)
		}
		return nil, fmt.Errorf("fehler beim Abrufen des Matches: %w", err)
	}

	return match, nil
}

// GetMatchesByDivision ruft alle Matches einer Division ab
func (d *Database) GetMatchesByDivision(division int) ([]*Match, error) {
	rows, err := d.DB.Query(
		`SELECT id, division, matchday, team_home_id, team_away_id, 
		 score_home, score_away, channel_id, reported_at, reported_by, created_at 
		 FROM matches WHERE division = ? ORDER BY matchday, id`,
		division,
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Matches: %w", err)
	}
	defer rows.Close()

	var matches []*Match
	for rows.Next() {
		match := &Match{}
		if err := rows.Scan(
			&match.ID, &match.Division, &match.Matchday, &match.TeamHomeID, &match.TeamAwayID,
			&match.ScoreHome, &match.ScoreAway, &match.ChannelID, &match.ReportedAt,
			&match.ReportedBy, &match.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("fehler beim Scannen des Matches: %w", err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// GetMatchesByDivisionAndMatchday ruft alle Matches eines Spieltags ab
func (d *Database) GetMatchesByDivisionAndMatchday(division, matchday int) ([]*Match, error) {
	rows, err := d.DB.Query(
		`SELECT id, division, matchday, team_home_id, team_away_id, 
		 score_home, score_away, channel_id, reported_at, reported_by, created_at 
		 FROM matches WHERE division = ? AND matchday = ? ORDER BY id`,
		division, matchday,
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Matches: %w", err)
	}
	defer rows.Close()

	var matches []*Match
	for rows.Next() {
		match := &Match{}
		if err := rows.Scan(
			&match.ID, &match.Division, &match.Matchday, &match.TeamHomeID, &match.TeamAwayID,
			&match.ScoreHome, &match.ScoreAway, &match.ChannelID, &match.ReportedAt,
			&match.ReportedBy, &match.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("fehler beim Scannen des Matches: %w", err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// UpdateMatchScore aktualisiert das Ergebnis eines Matches
func (d *Database) UpdateMatchScore(id, scoreHome, scoreAway int, reportedBy string) error {
	if scoreHome < 0 || scoreHome > 4 || scoreAway < 0 || scoreAway > 4 {
		return fmt.Errorf("scores müssen zwischen 0 und 4 liegen")
	}

	result, err := d.DB.Exec(
		`UPDATE matches 
		 SET score_home = ?, score_away = ?, reported_at = CURRENT_TIMESTAMP, reported_by = ? 
		 WHERE id = ?`,
		scoreHome, scoreAway, reportedBy, id,
	)
	if err != nil {
		return fmt.Errorf("fehler beim Aktualisieren des Scores: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fehler beim Überprüfen der aktualisierten Zeilen: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("match mit ID %d nicht gefunden", id)
	}

	return nil
}

// UpdateMatchChannelID aktualisiert die Channel-ID eines Matches
func (d *Database) UpdateMatchChannelID(id int, channelID string) error {
	result, err := d.DB.Exec(
		"UPDATE matches SET channel_id = ? WHERE id = ?",
		channelID, id,
	)
	if err != nil {
		return fmt.Errorf("fehler beim Aktualisieren der Channel-ID: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fehler beim Überprüfen der aktualisierten Zeilen: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("match mit ID %d nicht gefunden", id)
	}

	return nil
}

// DeleteMatchesByDivision löscht alle Matches einer Division
func (d *Database) DeleteMatchesByDivision(division int) error {
	_, err := d.DB.Exec("DELETE FROM matches WHERE division = ?", division)
	if err != nil {
		return fmt.Errorf("fehler beim Löschen der Matches: %w", err)
	}
	return nil
}

// GetMatchByChannelID ruft ein Match anhand der Channel-ID ab
func (d *Database) GetMatchByChannelID(channelID string) (*Match, error) {
	match := &Match{}
	err := d.DB.QueryRow(
		`SELECT id, division, matchday, team_home_id, team_away_id, 
		 score_home, score_away, channel_id, reported_at, reported_by, created_at 
		 FROM matches WHERE channel_id = ?`,
		channelID,
	).Scan(
		&match.ID, &match.Division, &match.Matchday, &match.TeamHomeID, &match.TeamAwayID,
		&match.ScoreHome, &match.ScoreAway, &match.ChannelID, &match.ReportedAt,
		&match.ReportedBy, &match.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("kein match für diesen channel gefunden")
		}
		return nil, fmt.Errorf("fehler beim Abrufen des Matches: %w", err)
	}

	return match, nil
}
