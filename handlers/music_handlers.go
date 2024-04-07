package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func helpCommand(h *Handler, s *discordgo.Session, m commandState, args ...string) {
	_, err := s.ChannelMessageSend(m.ChannelID, "```!verzoek <youtube link> - speel een liedje af\n!overslaan - sla het huidige liedje over\n!stoppen - stop met afspelen```")
	if err != nil {
		log.Fatalf("Error sending message: %s", err)
	}
}

func PlayCommand(h *Handler, s *discordgo.Session, m commandState, args ...string) {
	if len(args) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Ja ja ik kan maar 1 ding tegelijk doen!")
		if err != nil {
			log.Fatalf("Error sending message: %s", err)
		}

		return
	}

	var content string

	if strings.Contains(args[0], "list=") {
		_, err := s.ChannelMessageSend(m.ChannelID, "Zo dat zijn er veel die moet ik even opzoeken")
		videoTitles, err := h.queueManager.AddPlaylistToQueue(args[0])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Die liedjes ken ik niet")
			if err != nil {
				log.Fatalf("Error sending message: %s", err)
			}
			return
		}

		fmt.Println(videoTitles)

		content = "Ik heb echt heeeel veel liedjes toegevoegd"
	} else {
		video, err := h.queueManager.AddToQueue(args[0])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Dat liedje ken ik niet")
			if err != nil {
				log.Fatalf("Error sending message: %s", err)
			}
			return
		}

		content = video.Title
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Ha leuk die ken ik! komt er aan: ```%s```", content))

	if h.queueManager.CurrentlyPlaying {
		return
	}

	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.AuthorID {
			err = playSong(h, s, g.ID, vs.ChannelID)
			if err != nil {
				fmt.Println("Error playing sound:", err)
			}

			return
		}
	}
}

func playSong(h *Handler, s *discordgo.Session, guildID, channelID string) (err error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	time.Sleep(250 * time.Millisecond)
	vc.Speaking(true)

	h.queueManager.SetVoiceConnection(vc)
	h.queueManager.PlayQueue(s)

	vc.Speaking(false)

	time.Sleep(250 * time.Millisecond)

	vc.Disconnect()

	return nil
}

func SkipCommand(h *Handler, s *discordgo.Session, m commandState, args ...string) {
	h.queueManager.Skip()
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Nou als het zo moet speel ik hem ook niet af: ```%s```", h.queueManager.CurrentSong.Title))
	if err != nil {
		log.Fatalf("Error sending message: %s", err)
	}
}

func ClearCommand(h *Handler, s *discordgo.Session, m commandState, args ...string) {
	h.queueManager.ClearQueue()
	h.queueManager.Skip()
	_, err := s.ChannelMessageSend(m.ChannelID, "Ik heb de queue geleegd")
	if err != nil {
		log.Fatalf("Error sending message: %s", err)
	}
}
