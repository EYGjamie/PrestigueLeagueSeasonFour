package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jamie/prestigeleagueseasonfour/internal/bot"
	"github.com/jamie/prestigeleagueseasonfour/internal/database"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN Umgebungsvariable nicht gesetzt")
	}

	// Datenbank öffnen
	db, err := database.New("data/league.db")
	if err != nil {
		log.Fatalf("Fehler beim Öffnen der Datenbank: %v", err)
	}
	defer db.Close()

	fmt.Println("Datenbank verbunden!")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Fehler beim Erstellen der Discord Session: %v", err)
	}

	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	bot.SetDatabase(db)
	bot.RegisterHandlers(discord)

	err = discord.Open()
	if err != nil {
		log.Fatalf("Fehler beim Öffnen der Verbindung: %v", err)
	}
	defer discord.Close()

	fmt.Println("Bot läuft. Drücke CTRL+C zum Beenden.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
