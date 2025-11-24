package channels

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

// isGameFree prüft ob es ein spielfreies Match ist (Team ist NULL oder ID ist 0)
func isGameFree(team *database.Team) bool {
	return team == nil || team.ID == 0
}

// CreateMatchChannel erstellt einen Discord Channel für ein Match
func CreateMatchChannel(s *discordgo.Session, guildID, categoryID string, match *database.Match, homeTeam, awayTeam *database.Team) (string, error) {
	// Channel Name erstellen
	channelName := formatChannelName(match.Division, match.Matchday, homeTeam, awayTeam)

	// Permission Overwrites
	permissions := []*discordgo.PermissionOverwrite{
		// @everyone darf nichts sehen
		{
			ID:   guildID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
	}

	// Home Team Rolle hinzufügen (nur wenn kein Game-free)
	if !isGameFree(homeTeam) && homeTeam.RoleID != "" {
		permissions = append(permissions, &discordgo.PermissionOverwrite{
			ID:    homeTeam.RoleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		})
	}

	// Away Team Rolle hinzufügen (nur wenn kein Free Win)
	if !isGameFree(awayTeam) && awayTeam.RoleID != "" {
		permissions = append(permissions, &discordgo.PermissionOverwrite{
			ID:    awayTeam.RoleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		})
	}

	// Channel erstellen
	channel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             categoryID,
		PermissionOverwrites: permissions,
	})
	if err != nil {
		return "", fmt.Errorf("fehler beim Erstellen des Channels: %w", err)
	}

	// Willkommensnachricht senden
	if err := sendWelcomeMessage(s, channel.ID, homeTeam, awayTeam, match); err != nil {
		// Channel löschen bei Fehler
		s.ChannelDelete(channel.ID)
		return "", fmt.Errorf("fehler beim Senden der Willkommensnachricht: %w", err)
	}

	return channel.ID, nil
}

// formatChannelName erstellt den Channel-Namen
func formatChannelName(division, matchday int, homeTeam, awayTeam *database.Team) string {
	homeName := "Game-free"
	if !isGameFree(homeTeam) {
		homeName = homeTeam.Name
	}

	awayTeamName := "Game-free"
	if !isGameFree(awayTeam) {
		awayTeamName = awayTeam.Name
	}

	// Namen für Discord formatieren (lowercase, keine Leerzeichen)
	homeName = strings.ToLower(strings.ReplaceAll(homeName, " ", "-"))
	awayName := strings.ToLower(strings.ReplaceAll(awayTeamName, " ", "-"))

	// Sonderzeichen entfernen
	homeName = sanitizeChannelName(homeName)
	awayName = sanitizeChannelName(awayName)

	return fmt.Sprintf("div%d-woche%d-%s-%s", division, matchday, homeName, awayName)
}

// sanitizeChannelName entfernt ungültige Zeichen für Discord Channel-Namen
func sanitizeChannelName(name string) string {
	// Nur Buchstaben, Zahlen und Bindestriche erlaubt
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// sendWelcomeMessage sendet die Willkommensnachricht mit Pings
func sendWelcomeMessage(s *discordgo.Session, channelID string, homeTeam, awayTeam *database.Team, match *database.Match) error {
	// Rollen-Pings
	var pings []string
	if !isGameFree(homeTeam) && homeTeam.RoleID != "" {
		pings = append(pings, fmt.Sprintf("<@&%s>", homeTeam.RoleID))
	}
	if !isGameFree(awayTeam) && awayTeam.RoleID != "" {
		pings = append(pings, fmt.Sprintf("<@&%s>", awayTeam.RoleID))
	}

	pingMessage := strings.Join(pings, " ")

	// Ping-Nachricht senden
	if pingMessage != "" {
		_, err := s.ChannelMessageSend(channelID, pingMessage)
		if err != nil {
			return err
		}
	}

	// Embed erstellen
	var embed *discordgo.MessageEmbed

	// Prüfen, ob es ein spielfreies Match ist
	if isGameFree(awayTeam) || isGameFree(homeTeam) {
		// Welches Team hat spielfrei?
		teamName := ""
		if !isGameFree(homeTeam) {
			teamName = homeTeam.Name
		} else if !isGameFree(awayTeam) {
			teamName = awayTeam.Name
		} else {
			teamName = "Unbekanntes Team"
		}

		// Spielfrei-Nachricht
		embed = &discordgo.MessageEmbed{
			Title:       "🌴 Spielfreie Woche / Game-free Week",
			Description: fmt.Sprintf("**%s** hat in **Woche %d** spielfrei!\n**%s** has a bye in **Week %d**!", teamName, match.Matchday, teamName, match.Matchday),
			Color:       0x57F287, // Grün
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "ℹ️ Information",
					Value:  "Euer Team hat diese Woche kein Match.\nYour team has no match this week.",
					Inline: false,
				},
				{
					Name:   "💪 Training",
					Value:  "Nutzt die Zeit zum Trainieren und Verbessern!\nUse this time to train and improve!",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Viel Erfolg in der nächsten Woche! Good luck next week! 🚀",
			},
		}
	} else {
		// Normales Match
		awayTeamName := awayTeam.Name

		// Best-of Format basierend auf Division
		matchDetails := "Best of 5 (First to 3 wins)\nBest of 5 (Erster mit 3 Siegen)"
		if match.Division == 1 || match.Division == 2 {
			matchDetails = "Best of 7 (First to 4 wins)\nBest of 7 (Erster mit 4 Siegen)"
		}

		embed = &discordgo.MessageEmbed{
			Title:       "🏆 Match Information",
			Description: fmt.Sprintf("Willkommen zum Match der **Woche %d**!\nWelcome to the match of **Week %d**!", match.Matchday, match.Matchday),
			Color:       0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Home Team",
					Value:  fmt.Sprintf("**%s**", homeTeam.Name),
					Inline: true,
				},
				{
					Name:   "Away Team",
					Value:  fmt.Sprintf("**%s**", awayTeamName),
					Inline: true,
				},
				{
					Name:   "\u200b",
					Value:  "\u200b",
					Inline: false,
				},
				{
					Name:   "📋 Match Details / Spieldetails",
					Value:  matchDetails,
					Inline: false,
				},
				{
					Name:   "⚙️ Spielmodus / Game Mode",
					Value:  "3v3 Standard Competitive",
					Inline: false,
				},
				{
					Name:   "📅 Termin / Schedule",
					Value:  "Bitte koordiniert euren Spieltermin in diesem Channel.\nPlease coordinate your match date in this channel.",
					Inline: false,
				},
				{
					Name:   "📊 Ergebnis melden / Report Result",
					Value:  "Nach dem Match bitte das Ergebnis mit `/report_result` eintragen und ggf. Screenshots posten.\nAfter the match, please report the result with `/report_result` and post screenshots if necessary.",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Viel Erfolg beim Match! Good luck! 🚀",
			},
		}
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}
