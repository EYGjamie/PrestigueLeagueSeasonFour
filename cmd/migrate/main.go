package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

func main() {
	// Datenbank öffnen
	db, err := database.New("data/league.db")
	if err != nil {
		log.Fatalf("Fehler beim Öffnen der Datenbank: %v", err)
	}
	defer db.Close()

	fmt.Println("Datenbank erfolgreich initialisiert!")

	// CSV Datei einlesen (optional)
	if len(os.Args) > 1 && os.Args[1] == "--import" {
		// Rollen laden
		roles, err := loadRolesFromCSV("Data/roles.csv")
		if err != nil {
			log.Fatalf("Fehler beim Laden der Rollen: %v", err)
		}
		fmt.Printf("%d Rollen geladen\n", len(roles))

		if err := importTeamsFromCSV(db, "Data/teams.csv", roles); err != nil {
			log.Fatalf("Fehler beim Importieren der Teams: %v", err)
		}
		fmt.Println("Teams erfolgreich importiert!")
	}
}

func importTeamsFromCSV(db *database.Database, filePath string, roles map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("fehler beim Öffnen der CSV Datei: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	// Header überspringen
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("fehler beim Lesen des Headers: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("fehler beim Lesen der CSV Datei: %w", err)
	}

	for i, record := range records {
		if len(record) < 2 {
			continue
		}

		teamName := strings.TrimSpace(record[0])
		if teamName == "" {
			continue
		}

		// Division aus letzter Spalte
		divisionStr := strings.TrimSpace(record[len(record)-1])
		division, err := strconv.Atoi(divisionStr)
		if err != nil {
			log.Printf("Warnung: Ungültige Division für Team '%s': %v", teamName, err)
			continue
		}

		// Team erstellen
		team, err := db.CreateTeam(teamName, division)
		if err != nil {
			// Wenn Team bereits existiert, überspringen
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				log.Printf("Team '%s' existiert bereits, überspringe...", teamName)
				continue
			}
			return fmt.Errorf("fehler beim Erstellen des Teams '%s': %w", teamName, err)
		}

		// Rollen-ID setzen, falls vorhanden
		if roleID, exists := roles[teamName]; exists {
			if err := db.UpdateTeamRoleID(team.ID, roleID); err != nil {
				log.Printf("Warnung: Fehler beim Setzen der Rollen-ID für Team '%s': %v", teamName, err)
			} else {
				fmt.Printf("[%d/%d] Team erstellt: ID=%d, Name=%s, Division=%d, RoleID=%s\n",
					i+1, len(records), team.ID, team.Name, team.Division, roleID)
				continue
			}
		}

		fmt.Printf("[%d/%d] Team erstellt: ID=%d, Name=%s, Division=%d\n",
			i+1, len(records), team.ID, team.Name, team.Division)
	}

	return nil
}

func loadRolesFromCSV(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Rollen-CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	// Header überspringen
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("fehler beim Lesen des Headers: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der Rollen-CSV: %w", err)
	}

	roles := make(map[string]string)
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		roleName := strings.TrimSpace(record[0])
		roleID := strings.TrimSpace(record[1])
		if roleName != "" && roleID != "" {
			roles[roleName] = roleID
		}
	}

	return roles, nil
}
