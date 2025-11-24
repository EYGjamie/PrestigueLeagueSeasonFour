package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/channels"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

// CreateChannelsCommand erstellt Discord Channels für alle Matches aller Divisionen eines Matchdays
func CreateChannelsCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	options := i.ApplicationCommandData().Options
	if len(options) < 2 {
		respondError(s, i, "Bitte gib Matchday und Kategorie-ID an")
		return
	}

	matchday := int(options[0].IntValue())
	categoryID := options[1].StringValue()

	// Defer Antwort für längere Operationen
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Statistiken für alle Divisionen
	totalCreated := 0
	totalSkipped := 0
	totalErrors := 0
	divisionResults := make(map[int]string)

	// Alle Divisionen durchgehen (1-6)
	for division := 1; division <= 6; division++ {
		// Matches der Division und des Matchdays abrufen
		matches, err := db.GetMatchesByDivisionAndMatchday(division, matchday)
		if err != nil {
			divisionResults[division] = fmt.Sprintf("❌ Fehler: %v", err)
			continue
		}

		if len(matches) == 0 {
			divisionResults[division] = "⚠️ Keine Matches gefunden"
			continue
		}

		// Channels erstellen
		created := 0
		skipped := 0
		errors := 0

		for _, match := range matches {
			// Überspringen wenn bereits Channel existiert
			if match.ChannelID.Valid && match.ChannelID.String != "" {
				skipped++
				continue
			}

			// Teams abrufen
			homeTeam, err := db.GetTeamByID(match.TeamHomeID)
			if err != nil {
				errors++
				continue
			}

			var awayTeam *database.Team
			if match.TeamAwayID.Valid {
				awayTeam, err = db.GetTeamByID(int(match.TeamAwayID.Int64))
				if err != nil {
					errors++
					continue
				}
			}

			// Channel erstellen
			channelID, err := channels.CreateMatchChannel(s, i.GuildID, categoryID, match, homeTeam, awayTeam)
			if err != nil {
				errors++
				continue
			}

			// Channel-ID in Datenbank speichern
			if err := db.UpdateMatchChannelID(match.ID, channelID); err != nil {
				// Channel wieder löschen bei DB-Fehler
				s.ChannelDelete(channelID)
				errors++
				continue
			}

			created++
		}

		// Ergebnis für diese Division speichern
		if created > 0 {
			divisionResults[division] = fmt.Sprintf("✅ %d erstellt", created)
			if skipped > 0 {
				divisionResults[division] += fmt.Sprintf(", %d übersprungen", skipped)
			}
			if errors > 0 {
				divisionResults[division] += fmt.Sprintf(", ❌ %d Fehler", errors)
			}
		} else if skipped > 0 {
			divisionResults[division] = fmt.Sprintf("⏭️ Alle %d Channels bereits vorhanden", skipped)
		} else {
			divisionResults[division] = fmt.Sprintf("❌ %d Fehler", errors)
		}

		totalCreated += created
		totalSkipped += skipped
		totalErrors += errors
	}

	// Ergebnis-Embed mit Details pro Division
	embedFields := []*discordgo.MessageEmbedField{
		{
			Name:   "📊 Gesamtstatistik",
			Value:  fmt.Sprintf("✅ **%d** erstellt | ⏭️ **%d** übersprungen | ❌ **%d** Fehler", totalCreated, totalSkipped, totalErrors),
			Inline: false,
		},
	}

	// Details pro Division
	for div := 1; div <= 6; div++ {
		if result, ok := divisionResults[div]; ok {
			embedFields = append(embedFields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Division %d", div),
				Value:  result,
				Inline: true,
			})
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🏟️ Match Channels erstellt",
		Description: fmt.Sprintf("Channels für **Spieltag %d** aller Divisionen wurden erstellt.", matchday),
		Color:       0x00ff00,
		Fields:      embedFields,
	}

	if totalErrors > 0 {
		embed.Color = 0xffaa00
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Einige Channels konnten nicht erstellt werden. Prüfe die Bot-Berechtigungen.",
		}
	} else if totalCreated == 0 && totalSkipped > 0 {
		embed.Color = 0x0099ff
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Alle Channels waren bereits vorhanden.",
		}
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func followUpError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr("❌ " + message),
	})
}

func stringPtr(s string) *string {
	return &s
}
