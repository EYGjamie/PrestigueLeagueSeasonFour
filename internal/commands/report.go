package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

// ReportResultCommand öffnet ein Modal zum Eintragen des Ergebnisses
func ReportResultCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	// Match anhand Channel-ID abrufen
	match, err := db.GetMatchByChannelID(i.ChannelID)
	if err != nil {
		respondError(s, i, "Dieser Command kann nur in einem Match-Channel verwendet werden")
		return
	}

	// Teams abrufen
	homeTeam, err := db.GetTeamByID(match.TeamHomeID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Abrufen des Home Teams: %v", err))
		return
	}

	awayTeamName := "Free Win"
	if match.TeamAwayID.Valid {
		awayTeam, err := db.GetTeamByID(int(match.TeamAwayID.Int64))
		if err != nil {
			respondError(s, i, fmt.Sprintf("Fehler beim Abrufen des Away Teams: %v", err))
			return
		}
		awayTeamName = awayTeam.Name
	}

	// Best-of Format basierend auf Division
	placeholder := "0-3"
	if match.Division == 1 || match.Division == 2 {
		placeholder = "0-4"
	}

	// Modal erstellen
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("report_result:%d", match.ID),
			Title:    "Match Ergebnis eintragen",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "score_home",
							Label:       fmt.Sprintf("%s - Wins", homeTeam.Name),
							Style:       discordgo.TextInputShort,
							Placeholder: placeholder,
							Required:    true,
							MaxLength:   1,
							MinLength:   1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "score_away",
							Label:       fmt.Sprintf("%s - Wins", awayTeamName),
							Style:       discordgo.TextInputShort,
							Placeholder: placeholder,
							Required:    true,
							MaxLength:   1,
							MinLength:   1,
						},
					},
				},
			},
		},
	})

	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Öffnen des Modals: %v", err))
	}
}

// HandleReportResultModal verarbeitet das Modal Submit
func HandleReportResultModal(s *discordgo.Session, i *discordgo.InteractionCreate, db *database.Database) {
	data := i.ModalSubmitData()

	// Match-ID aus CustomID extrahieren
	var matchID int
	_, err := fmt.Sscanf(data.CustomID, "report_result:%d", &matchID)
	if err != nil {
		respondError(s, i, "Ungültige Modal-ID")
		return
	}

	// Match abrufen für Division-Check
	match, err := db.GetMatchByID(matchID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Match nicht gefunden: %v", err))
		return
	}

	// Best-of Format basierend auf Division
	maxScore := 3 // Best of 5
	if match.Division == 1 || match.Division == 2 {
		maxScore = 4 // Best of 7
	}

	// Scores aus Modal auslesen
	scoreHomeStr := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	scoreAwayStr := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	scoreHome, err := strconv.Atoi(scoreHomeStr)
	if err != nil || scoreHome < 0 || scoreHome > maxScore {
		respondError(s, i, fmt.Sprintf("Ungültiger Score für Home Team (muss 0-%d sein)", maxScore))
		return
	}

	scoreAway, err := strconv.Atoi(scoreAwayStr)
	if err != nil || scoreAway < 0 || scoreAway > maxScore {
		respondError(s, i, fmt.Sprintf("Ungültiger Score für Away Team (muss 0-%d sein)", maxScore))
		return
	}

	// Best-of validieren (einer muss maxScore haben)
	if scoreHome != maxScore && scoreAway != maxScore {
		respondError(s, i, fmt.Sprintf("Ein Team muss %d Wins haben", maxScore))
		return
	}

	if scoreHome == maxScore && scoreAway == maxScore {
		respondError(s, i, fmt.Sprintf("Beide Teams können nicht %d Wins haben", maxScore))
		return
	}

	// Teams abrufen
	homeTeam, err := db.GetTeamByID(match.TeamHomeID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Abrufen des Teams: %v", err))
		return
	}

	awayTeamName := "Free Win"
	if match.TeamAwayID.Valid {
		awayTeam, err := db.GetTeamByID(int(match.TeamAwayID.Int64))
		if err != nil {
			respondError(s, i, fmt.Sprintf("Fehler beim Abrufen des Teams: %v", err))
			return
		}
		awayTeamName = awayTeam.Name
	}

	// Ergebnis in Datenbank speichern
	reportedBy := i.Member.User.ID
	err = db.UpdateMatchScore(matchID, scoreHome, scoreAway, reportedBy)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Speichern des Ergebnisses: %v", err))
		return
	}

	// Bestätigung an User
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Ergebnis erfolgreich eingetragen! / Result successfully reported!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	// Öffentliche Benachrichtigung im Channel
	winner := homeTeam.Name
	if scoreAway > scoreHome {
		winner = awayTeamName
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📊 Match Ergebnis / Match Result",
		Description: "Das Match wurde abgeschlossen!\nThe match has been completed!",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   homeTeam.Name + " (Home)",
				Value:  fmt.Sprintf("**%d** Wins / Siege", scoreHome),
				Inline: true,
			},
			{
				Name:   awayTeamName + " (Away)",
				Value:  fmt.Sprintf("**%d** Wins / Siege", scoreAway),
				Inline: true,
			},
			{
				Name:   "\u200b",
				Value:  "\u200b",
				Inline: false,
			},
			{
				Name:   "🏆 Gewinner / Winner",
				Value:  fmt.Sprintf("**%s**", winner),
				Inline: false,
			},
			{
				Name:   "👤 Eingetragen von / Reported by",
				Value:  fmt.Sprintf("<@%s>", reportedBy),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.ChannelMessageSendEmbed(i.ChannelID, embed)
}
