package database

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

type Database struct {
	DB *sql.DB
}

// New erstellt eine neue Datenbankverbindung und initialisiert das Schema
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Datenbank: %w", err)
	}

	// Verbindung testen
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("fehler beim Verbinden zur Datenbank: %w", err)
	}

	database := &Database{DB: db}

	// Schema initialisieren
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("fehler beim Initialisieren des Schemas: %w", err)
	}

	return database, nil
}

// initSchema führt das SQL Schema aus
func (d *Database) initSchema() error {
	_, err := d.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("fehler beim Ausführen des Schemas: %w", err)
	}
	return nil
}

// Close schließt die Datenbankverbindung
func (d *Database) Close() error {
	return d.DB.Close()
}
