package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Team repräsentiert ein Team in der Datenbank
type Team struct {
	ID             int
	Name           string
	Division       int
	RoleID         string
	IsDisqualified bool
	DisqualifiedAt sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateTeam erstellt ein neues Team
func (d *Database) CreateTeam(name string, division int) (*Team, error) {
	result, err := d.DB.Exec(
		"INSERT INTO teams (name, division, role_id) VALUES (?, ?, ?)",
		name, division, "",
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Teams: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Team-ID: %w", err)
	}

	return d.GetTeamByID(int(id))
}

// GetTeamByID ruft ein Team anhand der ID ab
func (d *Database) GetTeamByID(id int) (*Team, error) {
	team := &Team{}
	err := d.DB.QueryRow(
		"SELECT id, name, division, role_id, is_disqualified, disqualified_at, created_at, updated_at FROM teams WHERE id = ?",
		id,
	).Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.IsDisqualified, &team.DisqualifiedAt, &team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team mit ID %d nicht gefunden", id)
		}
		return nil, fmt.Errorf("fehler beim Abrufen des Teams: %w", err)
	}

	return team, nil
}

// GetTeamByName ruft ein Team anhand des Namens ab
func (d *Database) GetTeamByName(name string) (*Team, error) {
	team := &Team{}
	err := d.DB.QueryRow(
		"SELECT id, name, division, role_id, is_disqualified, disqualified_at, created_at, updated_at FROM teams WHERE name = ?",
		name,
	).Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.IsDisqualified, &team.DisqualifiedAt, &team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team '%s' nicht gefunden", name)
		}
		return nil, fmt.Errorf("fehler beim Abrufen des Teams: %w", err)
	}

	return team, nil
}

// GetAllTeams ruft alle Teams ab
func (d *Database) GetAllTeams() ([]*Team, error) {
	rows, err := d.DB.Query(
		"SELECT id, name, division, role_id, is_disqualified, disqualified_at, created_at, updated_at FROM teams ORDER BY division, name",
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.IsDisqualified, &team.DisqualifiedAt, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, fmt.Errorf("fehler beim Scannen des Teams: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// GetTeamsByDivision ruft alle Teams einer Division ab
func (d *Database) GetTeamsByDivision(division int) ([]*Team, error) {
	rows, err := d.DB.Query(
		"SELECT id, name, division, role_id, is_disqualified, disqualified_at, created_at, updated_at FROM teams WHERE division = ? ORDER BY name",
		division,
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.IsDisqualified, &team.DisqualifiedAt, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, fmt.Errorf("fehler beim Scannen des Teams: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// UpdateTeam aktualisiert ein Team
func (d *Database) UpdateTeam(id int, name string, division int) error {
	result, err := d.DB.Exec(
		"UPDATE teams SET name = ?, division = ? WHERE id = ?",
		name, division, id,
	)
	if err != nil {
		return fmt.Errorf("fehler beim Aktualisieren des Teams: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fehler beim Überprüfen der aktualisierten Zeilen: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("team mit ID %d nicht gefunden", id)
	}

	return nil
}

// UpdateTeamRoleID aktualisiert die Discord Rollen-ID eines Teams
func (d *Database) UpdateTeamRoleID(id int, roleID string) error {
	result, err := d.DB.Exec(
		"UPDATE teams SET role_id = ? WHERE id = ?",
		roleID, id,
	)
	if err != nil {
		return fmt.Errorf("fehler beim Aktualisieren der Rollen-ID: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fehler beim Überprüfen der aktualisierten Zeilen: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("team mit ID %d nicht gefunden", id)
	}

	return nil
}

// DeleteTeam löscht ein Team
func (d *Database) DeleteTeam(id int) error {
	result, err := d.DB.Exec("DELETE FROM teams WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("fehler beim Löschen des Teams: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fehler beim Überprüfen der gelöschten Zeilen: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("team mit ID %d nicht gefunden", id)
	}

	return nil
}

// GetTeamByRoleID ruft ein Team anhand der Discord Rollen-ID ab
func (d *Database) GetTeamByRoleID(roleID string) (*Team, error) {
	team := &Team{}
	err := d.DB.QueryRow(
		"SELECT id, name, division, role_id, is_disqualified, disqualified_at, created_at, updated_at FROM teams WHERE role_id = ?",
		roleID,
	).Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.IsDisqualified, &team.DisqualifiedAt, &team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team mit Rollen-ID %s nicht gefunden", roleID)
		}
		return nil, fmt.Errorf("fehler beim Abrufen des Teams: %w", err)
	}

	return team, nil
}

// DisqualifyTeam disqualifiziert ein Team und setzt alle Matches auf 0:3
func (d *Database) DisqualifyTeam(teamID int) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return fmt.Errorf("fehler beim Starten der Transaktion: %w", err)
	}
	defer tx.Rollback()

	// Team als disqualified markieren
	_, err = tx.Exec(
		"UPDATE teams SET is_disqualified = 1, disqualified_at = CURRENT_TIMESTAMP WHERE id = ?",
		teamID,
	)
	if err != nil {
		return fmt.Errorf("fehler beim Disqualifizieren des Teams: %w", err)
	}

	// Alle Matches mit diesem Team auf 0:3 setzen
	// Home Matches (Team verliert 0:3)
	_, err = tx.Exec(`
		UPDATE matches 
		SET score_home = 0, score_away = 3, 
		    reported_at = CURRENT_TIMESTAMP, 
		    reported_by = 'System (Disqualified)'
		WHERE team_home_id = ? AND (score_home IS NULL OR score_away IS NULL)
	`, teamID)
	if err != nil {
		return fmt.Errorf("fehler beim Aktualisieren der Home Matches: %w", err)
	}

	// Away Matches (Team verliert 3:0)
	_, err = tx.Exec(`
		UPDATE matches 
		SET score_home = 3, score_away = 0,
		    reported_at = CURRENT_TIMESTAMP,
		    reported_by = 'System (Disqualified)'
		WHERE team_away_id = ? AND (score_home IS NULL OR score_away IS NULL)
	`, teamID)
	if err != nil {
		return fmt.Errorf("fehler beim Aktualisieren der Away Matches: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("fehler beim Commit der Transaktion: %w", err)
	}

	return nil
}

// RequalifyTeam hebt die Disqualifikation auf (setzt Matches NICHT zurück)
func (d *Database) RequalifyTeam(teamID int) error {
	result, err := d.DB.Exec(
		"UPDATE teams SET is_disqualified = 0, disqualified_at = NULL WHERE id = ?",
		teamID,
	)
	if err != nil {
		return fmt.Errorf("fehler beim Requalifizieren des Teams: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fehler beim Überprüfen der aktualisierten Zeilen: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("team mit ID %d nicht gefunden", teamID)
	}

	return nil
}
