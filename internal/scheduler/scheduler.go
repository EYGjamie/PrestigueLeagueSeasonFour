package scheduler

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed schedules.json
var schedulesJSON []byte

type Matchday struct {
	Home []int `json:"home"`
	Away []int `json:"away"`
}

type Schedules struct {
	Eight []Matchday `json:"8"`
	Ten   []Matchday `json:"10"`
}

var schedules Schedules

func init() {
	if err := json.Unmarshal(schedulesJSON, &schedules); err != nil {
		panic(fmt.Sprintf("fehler beim Laden der Spielpläne: %v", err))
	}
}

// GetSchedule gibt den passenden Spielplan für die Anzahl der Teams zurück
func GetSchedule(teamCount int) ([]Matchday, error) {
	switch {
	case teamCount <= 8:
		return schedules.Eight, nil
	case teamCount == 9, teamCount == 10:
		return schedules.Ten, nil
	default:
		return nil, fmt.Errorf("keine Spielplan-Vorlage für %d Teams verfügbar", teamCount)
	}
}

// GenerateMatches erstellt Match-Paarungen für eine Division
func GenerateMatches(teamIDs []int) ([][]Match, error) {
	teamCount := len(teamIDs)
	schedule, err := GetSchedule(teamCount)
	if err != nil {
		return nil, err
	}

	// Für 9 Teams: Letztes Team-Slot ist "Free Win"
	if teamCount == 9 {
		teamIDs = append(teamIDs, 0) // 0 = Free Win
	}

	var allMatchdays [][]Match
	for matchdayNum, matchday := range schedule {
		var matches []Match
		for i := 0; i < len(matchday.Home); i++ {
			homeIdx := matchday.Home[i] - 1
			awayIdx := matchday.Away[i] - 1

			if homeIdx >= len(teamIDs) || awayIdx >= len(teamIDs) {
				continue
			}

			homeTeamID := teamIDs[homeIdx]
			awayTeamID := teamIDs[awayIdx]

			matches = append(matches, Match{
				Matchday:   matchdayNum + 1,
				TeamHomeID: homeTeamID,
				TeamAwayID: awayTeamID,
			})
		}
		allMatchdays = append(allMatchdays, matches)
	}

	return allMatchdays, nil
}

// Match repräsentiert ein geplantes Spiel
type Match struct {
	Matchday   int
	TeamHomeID int
	TeamAwayID int // 0 = Free Win
}
