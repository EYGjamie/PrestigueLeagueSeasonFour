package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/commands"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

var db *database.Database

func SetDatabase(database *database.Database) {
	db = database
}

func RegisterHandlers(s *discordgo.Session) {
	s.AddHandler(messageCreate)
	s.AddHandler(ready)
	s.AddHandler(interactionCreate)
	s.AddHandler(modalSubmit)
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Liga Verwaltung")

	// Slash Commands registrieren
	registerCommands(s)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Hier können Commands hinzugefügt werden
}

func registerCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:                     "schedule",
			Description:              "Erstellt einen Spielplan für eine Division",
			DefaultMemberPermissions: &[]int64{discordgo.PermissionAdministrator}[0],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "division",
					Description: "Die Division für den Spielplan",
					Required:    true,
				},
			},
		},
		{
			Name:                     "createchannels",
			Description:              "Erstellt Discord Channels für alle Matches aller Divisionen eines Spieltags",
			DefaultMemberPermissions: &[]int64{discordgo.PermissionAdministrator}[0],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "matchday",
					Description: "Der Spieltag für den die Channels erstellt werden",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Die Kategorie-ID unter der die Channels erstellt werden",
					Required:    true,
				},
			},
		},
		{
			Name:        "report_result",
			Description: "Trägt das Ergebnis eines Matches ein (nur in Match-Channels)",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			panic(err)
		}
	}
}

// hasAdminPermission prüft ob der User Administrator-Rechte hat
func hasAdminPermission(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	// Spezielle User-ID die immer berechtigt ist
	if i.Member != nil && i.Member.User.ID == "423480294948208661" {
		return true
	}

	// Guild Member Permissions prüfen
	member := i.Member
	if member == nil {
		return false
	}

	// Prüfe ob der User Administrator-Rechte hat
	permissions := member.Permissions
	return (permissions & discordgo.PermissionAdministrator) != 0
}

// respondError sendet eine Fehlermeldung als Antwort
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "schedule":
		if !hasAdminPermission(s, i) {
			respondError(s, i, "❌ Dieser Command kann nur von Administratoren ausgeführt werden.")
			return
		}
		commands.ScheduleCommand(s, i, db)
	case "createchannels":
		if !hasAdminPermission(s, i) {
			respondError(s, i, "❌ Dieser Command kann nur von Administratoren ausgeführt werden.")
			return
		}
		commands.CreateChannelsCommand(s, i, db)
	case "report_result":
		commands.ReportResultCommand(s, i, db)
	}
}

func modalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}

	// Report Result Modal
	if len(i.ModalSubmitData().CustomID) > 14 && i.ModalSubmitData().CustomID[:14] == "report_result:" {
		commands.HandleReportResultModal(s, i, db)
	}
}
