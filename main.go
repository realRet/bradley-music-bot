package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"music-bot-go/handlers"
)

var handler = handlers.NewHandler("!")

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(interactionCreate)

	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentGuildVoiceStates

	err = discord.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %s", err)
	}

	token := os.Getenv("GUILD_ID")

	handler.CreateCommands(discord, token)

	sc := make(chan os.Signal, 1)
	<-sc

	discord.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "/verzoek <youtube link>")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handler.HandleTextCommand(s, m)
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handler.HandleSlashCommand(s, i)
}
