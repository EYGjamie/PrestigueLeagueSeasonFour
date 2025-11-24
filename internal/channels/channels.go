package channels

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

// CreateMatchChannel erstellt einen Discord Channel für ein Match
func CreateMatchChannel(s *discordgo.Session, guildID, categoryID string, match *database.Match, homeTeam, awayTeam *database.Team) (string, error) {
	// Channel Name erstellen
	channelName := formatChannelName(match.Division, match.Matchday, homeTeam.Name, awayTeam)

	// Permission Overwrites
	permissions := []*discordgo.PermissionOverwrite{
		// @everyone darf nichts sehen
		{
			ID:   guildID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
	}

	// Home Team Rolle hinzufügen
	if homeTeam.RoleID != "" {
		permissions = append(permissions, &discordgo.PermissionOverwrite{
			ID:    homeTeam.RoleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		})
	}

	// Away Team Rolle hinzufügen (nur wenn kein Free Win)
	if awayTeam != nil && awayTeam.RoleID != "" {
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
func formatChannelName(division, matchday int, homeTeamName string, awayTeam *database.Team) string {
	awayTeamName := "FreeWin"
	if awayTeam != nil {
		awayTeamName = awayTeam.Name
	}

	// Namen für Discord formatieren (lowercase, keine Leerzeichen)
	homeName := strings.ToLower(strings.ReplaceAll(homeTeamName, " ", "-"))
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
	if homeTeam.RoleID != "" {
		pings = append(pings, fmt.Sprintf("<@&%s>", homeTeam.RoleID))
	}
	if awayTeam != nil && awayTeam.RoleID != "" {
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
	awayTeamName := "Free Win"
	if awayTeam != nil {
		awayTeamName = awayTeam.Name
	}

	// Best-of Format basierend auf Division
	matchDetails := "Best of 5 (First to 3 wins)\nBest of 5 (Erster mit 3 Siegen)"
	if match.Division == 1 || match.Division == 2 {
		matchDetails = "Best of 7 (First to 4 wins)\nBest of 7 (Erster mit 4 Siegen)"
	}

	embed := &discordgo.MessageEmbed{
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

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}
