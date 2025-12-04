package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

// DisqualifyCommand disqualifiziert ein Team über seine Discord-Rolle
func DisqualifyCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	roleID := optionMap["team"].RoleValue(nil, "").ID

	// Team anhand der Rolle finden
	team, err := db.GetTeamByRoleID(roleID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Team mit dieser Rolle nicht gefunden: %v", err))
		return
	}

	// Prüfen, ob Team bereits disqualified ist
	if team.IsDisqualified {
		respondError(s, i, fmt.Sprintf("Team **%s** ist bereits disqualifiziert", team.Name))
		return
	}

	// Team disqualifizieren
	err = db.DisqualifyTeam(team.ID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Disqualifizieren des Teams: %v", err))
		return
	}

	// Erfolgs-Embed erstellen
	embed := &discordgo.MessageEmbed{
		Title:       "⚠️ Team Disqualifiziert",
		Description: fmt.Sprintf("**%s** wurde disqualifiziert.", team.Name),
		Color:       0xFF0000, // Rot
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Auswirkungen",
				Value:  "• Alle Matches werden mit 0:3 gegen das Team gewertet\n• Zukünftige Match-Channels werden als Free Win erstellt\n• Das Team wird in der Tabelle durchgestrichen angezeigt",
				Inline: false,
			},
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		log.Printf("Fehler beim Senden der Disqualify-Antwort: %v", err)
	}
}

// RequalifyCommand hebt die Disqualifikation eines Teams auf
func RequalifyCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	roleID := optionMap["team"].RoleValue(nil, "").ID

	// Team anhand der Rolle finden
	team, err := db.GetTeamByRoleID(roleID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Team mit dieser Rolle nicht gefunden: %v", err))
		return
	}

	// Prüfen, ob Team disqualified ist
	if !team.IsDisqualified {
		respondError(s, i, fmt.Sprintf("Team **%s** ist nicht disqualifiziert", team.Name))
		return
	}

	// Disqualifikation aufheben
	err = db.RequalifyTeam(team.ID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Requalifizieren des Teams: %v", err))
		return
	}

	// Erfolgs-Embed erstellen
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Disqualifikation aufgehoben",
		Description: fmt.Sprintf("**%s** ist nicht mehr disqualifiziert.", team.Name),
		Color:       0x00FF00, // Grün
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Hinweis",
				Value:  "Bereits gewertete Matches (0:3) werden NICHT automatisch zurückgesetzt.",
				Inline: false,
			},
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		log.Printf("Fehler beim Senden der Requalify-Antwort: %v", err)
	}
}
