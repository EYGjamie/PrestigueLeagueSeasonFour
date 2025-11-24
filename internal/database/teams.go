package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Team repräsentiert ein Team in der Datenbank
type Team struct {
	ID        int
	Name      string
	Division  int
	RoleID    string
	CreatedAt time.Time
	UpdatedAt time.Time
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
		"SELECT id, name, division, role_id, created_at, updated_at FROM teams WHERE id = ?",
		id,
	).Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.CreatedAt, &team.UpdatedAt)

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
		"SELECT id, name, division, role_id, created_at, updated_at FROM teams WHERE name = ?",
		name,
	).Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.CreatedAt, &team.UpdatedAt)

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
		"SELECT id, name, division, role_id, created_at, updated_at FROM teams ORDER BY division, name",
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, fmt.Errorf("fehler beim Scannen des Teams: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// GetTeamsByDivision ruft alle Teams einer Division ab
func (d *Database) GetTeamsByDivision(division int) ([]*Team, error) {
	rows, err := d.DB.Query(
		"SELECT id, name, division, role_id, created_at, updated_at FROM teams WHERE division = ? ORDER BY name",
		division,
	)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.Division, &team.RoleID, &team.CreatedAt, &team.UpdatedAt); err != nil {
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
