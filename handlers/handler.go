package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"music-bot-go/music"
)

type HandlerFunc func(h *Handler, s *discordgo.Session, m commandState, args ...string)

type commandState struct {
	ChannelID string
	AuthorID  string
	GuildID   string
}

type Handler struct {
	handlers        map[string]command
	queueManager    *music.QueueManager
	command_prefix  string
	discord_session *discordgo.Session
}

func NewHandler(prefix string) *Handler {
	handler := Handler{
		handlers:       make(map[string]command),
		queueManager:   music.NewQueueManager(),
		command_prefix: prefix,
	}

	// register commands here
	handler.AddCommand(newCommand("verzoek", "Speel een liedje af", PlayCommand, "query"))
	handler.AddCommand(newCommand("overslaan", "Sla het huidige liedje over", SkipCommand))
	handler.AddCommand(newCommand("stoppen", "Stop met afspelen", ClearCommand))
	handler.AddCommand(newCommand("help", "Laat alle commando's zien", helpCommand))

	return &handler
}

func (h *Handler) SetDiscordSession(s *discordgo.Session) {
	h.discord_session = s
}

func (h *Handler) AddCommand(c command) {
	h.handlers[c.name] = c
}

func (h *Handler) CreateCommands(guildID string) {
	registeredCommands, err := h.discord_session.ApplicationCommands(h.discord_session.State.User.ID, guildID)
	if err != nil {
		log.Fatalf("Could not fetch registered commands: %v", err)
	}

	for _, v := range registeredCommands {
		err := h.discord_session.ApplicationCommandDelete(h.discord_session.State.User.ID, guildID, v.ID)
		fmt.Println("Deleting command: " + v.Name)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	for _, c := range h.handlers {
		_, err := h.discord_session.ApplicationCommandCreate(h.discord_session.State.User.ID, guildID, c.GetCommand())
		if err != nil {
			fmt.Println(err)
			fmt.Println("Error creating command: " + c.name)
		}
	}
}

func (h *Handler) HandleTextCommand(m *discordgo.MessageCreate) {
	if m.Author.ID == h.discord_session.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, h.command_prefix) {
		fmt.Println("Handling message: " + m.Content)
		commands := strings.Split(m.Content, " ")
		command := strings.TrimPrefix(commands[0], h.command_prefix)

		if handler, ok := h.handlers[command]; ok {

			var args []string
			if len(commands) > 1 {
				args = commands[1:]
			}

			handler.runTextCommand(h, h.discord_session, m, args...)
		}
	}
}

func (h *Handler) HandleSlashCommand(i *discordgo.InteractionCreate) {
	if handler, ok := h.handlers[i.ApplicationCommandData().Name]; ok {
		go handler.runSlashCommand(h, h.discord_session, i)
		err := h.discord_session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: " ",
			},
		})
		if err != nil {
			fmt.Println("Error responding to interaction: ", err)
		}
		h.discord_session.InteractionResponseDelete(i.Interaction)
	}
}
