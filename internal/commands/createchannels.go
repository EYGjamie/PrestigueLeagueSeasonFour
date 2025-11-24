package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/channels"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

// CreateChannelsCommand erstellt Discord Channels für alle Matches einer Division und eines Matchdays
func CreateChannelsCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	options := i.ApplicationCommandData().Options
	if len(options) < 3 {
		respondError(s, i, "Bitte gib Division, Matchday und Kategorie-ID an")
		return
	}

	division := int(options[0].IntValue())
	matchday := int(options[1].IntValue())
	categoryID := options[2].StringValue()

	// Matches der Division und des Matchdays abrufen
	matches, err := db.GetMatchesByDivisionAndMatchday(division, matchday)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Abrufen der Matches: %v", err))
		return
	}

	if len(matches) == 0 {
		respondError(s, i, "Keine Matches für diese Division und Spieltag gefunden")
		return
	}

	// Defer Antwort für längere Operationen
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Channels erstellen
	created := 0
	skipped := 0
	errors := 0
	var errorLog []string

	for _, match := range matches {
		// Überspringen wenn bereits Channel existiert
		if match.ChannelID.Valid && match.ChannelID.String != "" {
			skipped++
			continue
		}

		// Teams abrufen
		var homeTeam *database.Team
		if match.TeamHomeID != 0 {
			homeTeam, err = db.GetTeamByID(match.TeamHomeID)
			if err != nil {
				errors++
				errorMsg := fmt.Sprintf("Match ID %d: Fehler beim Abrufen des Heimteams (ID %d): %v", match.ID, match.TeamHomeID, err)
				log.Println("[CreateChannels]", errorMsg)
				errorLog = append(errorLog, errorMsg)
				continue
			}
		}

		var awayTeam *database.Team
		if match.TeamAwayID.Valid && match.TeamAwayID.Int64 != 0 {
			awayTeam, err = db.GetTeamByID(int(match.TeamAwayID.Int64))
			if err != nil {
				errors++
				errorMsg := fmt.Sprintf("Match ID %d: Fehler beim Abrufen des Auswärtsteams (ID %d): %v", match.ID, match.TeamAwayID.Int64, err)
				log.Println("[CreateChannels]", errorMsg)
				errorLog = append(errorLog, errorMsg)
				continue
			}
		}

		// Überspringe, wenn beide Teams NULL/0 sind
		if homeTeam == nil && awayTeam == nil {
			skipped++
			continue
		}

		// Channel erstellen
		channelID, err := channels.CreateMatchChannel(s, i.GuildID, categoryID, match, homeTeam, awayTeam)
		if err != nil {
			errors++
			matchName := "Unknown vs Unknown"
			if homeTeam != nil && awayTeam != nil {
				matchName = fmt.Sprintf("%s vs %s", homeTeam.Name, awayTeam.Name)
			} else if homeTeam != nil {
				matchName = fmt.Sprintf("%s vs Game-free", homeTeam.Name)
			} else if awayTeam != nil {
				matchName = fmt.Sprintf("Game-free vs %s", awayTeam.Name)
			}
			errorMsg := fmt.Sprintf("Match ID %d (%s): Fehler beim Erstellen des Discord-Channels: %v", match.ID, matchName, err)
			log.Println("[CreateChannels]", errorMsg)
			errorLog = append(errorLog, errorMsg)
			continue
		}

		// Channel-ID in Datenbank speichern
		if err := db.UpdateMatchChannelID(match.ID, channelID); err != nil {
			// Channel wieder löschen bei DB-Fehler
			if _, delErr := s.ChannelDelete(channelID); delErr != nil {
				log.Printf("[CreateChannels] Match ID %d: Channel %s konnte nicht gelöscht werden: %v", match.ID, channelID, delErr)
			}
			errors++
			matchName := "Unknown vs Unknown"
			if homeTeam != nil && awayTeam != nil {
				matchName = fmt.Sprintf("%s vs %s", homeTeam.Name, awayTeam.Name)
			} else if homeTeam != nil {
				matchName = fmt.Sprintf("%s vs Game-free", homeTeam.Name)
			} else if awayTeam != nil {
				matchName = fmt.Sprintf("Game-free vs %s", awayTeam.Name)
			}
			errorMsg := fmt.Sprintf("Match ID %d (%s): Fehler beim Speichern der Channel-ID in DB: %v", match.ID, matchName, err)
			log.Println("[CreateChannels]", errorMsg)
			errorLog = append(errorLog, errorMsg)
			continue
		}

		created++
	}

	// Ergebnis-Embed
	embed := &discordgo.MessageEmbed{
		Title:       "🏟️ Match Channels erstellt",
		Description: fmt.Sprintf("Channels für **Division %d - Spieltag %d** wurden erstellt.", division, matchday),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "✅ Erstellt",
				Value:  fmt.Sprintf("%d Channels", created),
				Inline: true,
			},
			{
				Name:   "⏭️ Übersprungen",
				Value:  fmt.Sprintf("%d bereits vorhanden", skipped),
				Inline: true,
			},
			{
				Name:   "❌ Fehler",
				Value:  fmt.Sprintf("%d Fehler", errors),
				Inline: true,
			},
		},
	}

	// Fehlerdetails hinzufügen wenn vorhanden
	if len(errorLog) > 0 {
		// Begrenze auf die ersten 5 Fehler für Discord Embed
		errorDetails := ""
		maxErrors := 5
		if len(errorLog) > maxErrors {
			for i := 0; i < maxErrors; i++ {
				errorDetails += fmt.Sprintf("• %s\n", errorLog[i])
			}
			errorDetails += fmt.Sprintf("\n...und %d weitere Fehler (siehe Server-Logs)", len(errorLog)-maxErrors)
		} else {
			for _, errMsg := range errorLog {
				errorDetails += fmt.Sprintf("• %s\n", errMsg)
			}
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📋 Fehlerdetails",
			Value:  errorDetails,
			Inline: false,
		})
	}

	if errors > 0 {
		embed.Color = 0xffaa00
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Einige Channels konnten nicht erstellt werden. Prüfe die Bot-Berechtigungen und Server-Logs.",
		}
	} else if created == 0 && skipped > 0 {
		embed.Color = 0x0099ff
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Alle Channels waren bereits vorhanden.",
		}
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}
