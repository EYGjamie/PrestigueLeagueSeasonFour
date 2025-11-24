package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
	"github.com/jamie/prestigeleagueseasonfour/internal/scheduler"
)

// ScheduleCommand erstellt einen Spielplan für eine Division
func ScheduleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondError(s, i, "Bitte gib eine Division an")
		return
	}

	division := int(options[0].IntValue())

	// Teams der Division abrufen
	teams, err := db.GetTeamsByDivision(division)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Abrufen der Teams: %v", err))
		return
	}

	if len(teams) == 0 {
		respondError(s, i, fmt.Sprintf("Keine Teams in Division %d gefunden", division))
		return
	}

	if len(teams) < 3 {
		respondError(s, i, fmt.Sprintf("Division %d hat zu wenige Teams (%d). Mindestens 3 Teams benötigt.", division, len(teams)))
		return
	}

	// Team-IDs extrahieren
	var teamIDs []int
	teamNames := make(map[int]string)
	for _, team := range teams {
		teamIDs = append(teamIDs, team.ID)
		teamNames[team.ID] = team.Name
	}

	// Spielplan generieren
	matchdays, err := scheduler.GenerateMatches(teamIDs)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Generieren des Spielplans: %v", err))
		return
	}

	// Alte Matches löschen
	if err := db.DeleteMatchesByDivision(division); err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Löschen alter Matches: %v", err))
		return
	}

	// Matches in Datenbank speichern
	totalMatches := 0
	for _, matchday := range matchdays {
		for _, match := range matchday {
			var awayID *int
			if match.TeamAwayID != 0 {
				awayID = &match.TeamAwayID
			}

			_, err := db.CreateMatch(division, match.Matchday, match.TeamHomeID, awayID)
			if err != nil {
				respondError(s, i, fmt.Sprintf("Fehler beim Erstellen des Matches: %v", err))
				return
			}
			totalMatches++
		}
	}

	// Response erstellen
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Spielplan für Division %d erstellt", division),
		Description: fmt.Sprintf("**%d Teams**, **%d Spieltage**, **%d Matches**\n\n", len(teams), len(matchdays), totalMatches),
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Erste 3 Spieltage als Preview
	for mdIdx, matchday := range matchdays {
		if mdIdx >= 3 {
			break
		}

		var matchList []string
		for _, match := range matchday {
			homeName := teamNames[match.TeamHomeID]
			awayName := "Free Win"
			if match.TeamAwayID != 0 {
				awayName = teamNames[match.TeamAwayID]
			}
			matchList = append(matchList, fmt.Sprintf("• %s vs %s", homeName, awayName))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Spieltag %d", mdIdx+1),
			Value:  strings.Join(matchList, "\n"),
			Inline: false,
		})
	}

	if len(matchdays) > 3 {
		embed.Description += fmt.Sprintf("*... und %d weitere Spieltage*", len(matchdays)-3)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
