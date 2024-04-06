package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type command struct {
	name        string
	description string
	handler     HandlerFunc
	options     []string
	commandID   string
}

func (c *command) runTextCommand(h *Handler, s *discordgo.Session, m *discordgo.MessageCreate, args ...string) {
	commandState := commandState{
		ChannelID: m.ChannelID,
		AuthorID:  m.Author.ID,
		GuildID:   m.GuildID,
	}

	fmt.Println("Running text command: " + c.name + " with args: " + fmt.Sprint(args))

	c.handler(h, s, commandState, args...)
}

func (c *command) runSlashCommand(h *Handler, s *discordgo.Session, i *discordgo.InteractionCreate) {
	args := make([]string, 0)
	for _, v := range i.ApplicationCommandData().Options {
		args = append(args, v.StringValue())
	}

	commandState := commandState{
		ChannelID: i.ChannelID,
		AuthorID:  i.Member.User.ID,
		GuildID:   i.GuildID,
	}

	c.handler(h, s, commandState, args...)
}

func (c *command) GetCommand() *discordgo.ApplicationCommand {
	fmt.Println("Creating command: " + c.name)
	return &discordgo.ApplicationCommand{
		Name:        c.name,
		Description: c.description,
		Options:     c.GetOptions(),
	}
}

func (c *command) GetOptions() []*discordgo.ApplicationCommandOption {
	options := make([]*discordgo.ApplicationCommandOption, 0)
	for _, v := range c.options {
		options = append(options, &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        v,
			Description: c.description,
			Required:    true,
		})
	}

	return options
}

func newCommand(name string, description string, handler HandlerFunc, options ...string) command {
	return command{
		name:        name,
		description: description,
		handler:     handler,
		options:     options,
	}
}
