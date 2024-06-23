package music

import (
	"context"
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
	"github.com/ybaac/dca"
)

type QueueManager struct {
	Queue            []*youtube.Video
	CurrentSong      *youtube.Video
	CurrentlyPlaying bool
	voiceConnection  *discordgo.VoiceConnection
	Paused           bool
	skipChan         chan bool
	currentSongCtx   context.Context
	cancelSong       context.CancelFunc
	youtubeClient    *youtube.Client
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		Queue:            make([]*youtube.Video, 0),
		voiceConnection:  nil,
		CurrentlyPlaying: false,
		Paused:           false,
		skipChan:         make(chan bool),
		youtubeClient:    &youtube.Client{},
	}
}

func (q *QueueManager) AddToQueue(song string) (*youtube.Video, error) {
	video, err := q.youtubeClient.GetVideo(song)
	if err != nil {
		return nil, err
	}

	q.Queue = append(q.Queue, video)
	return video, nil
}

func (q *QueueManager) AddPlaylistToQueue(playlistUri string) (string, error) {
	playlist, err := q.youtubeClient.GetPlaylist(playlistUri)
	if err != nil {
		return "", err
	}

	var videoTitles string

	for _, entry := range playlist.Videos {
		video, err := q.youtubeClient.VideoFromPlaylistEntry(entry)
		if err != nil {
			panic(err)
		}

		q.Queue = append(q.Queue, video)
		videoTitles += video.Title + "\n"
	}

	return videoTitles, nil
}

func (q *QueueManager) NextSong() *youtube.Video {
	if len(q.Queue) == 0 {
		return nil
	}

	song := q.Queue[0]
	q.Queue = q.Queue[1:]

	return song
}

func (q *QueueManager) ClearQueue() {
	q.Queue = make([]*youtube.Video, 0)
}

func (q *QueueManager) IsEmpty() bool {
	return len(q.Queue) == 0
}

func (q *QueueManager) SetVoiceConnection(vc *discordgo.VoiceConnection) {
	q.voiceConnection = vc
}

func (q *QueueManager) Skip() {
	q.skipChan <- true
}

func (q *QueueManager) PlayQueue(session *discordgo.Session) {
	for !q.IsEmpty() {
		if !q.CurrentlyPlaying {
			song := q.NextSong()
			q.CurrentSong = song

			// Create a new context for each song
			ctx, cancel := context.WithCancel(context.Background())
			q.currentSongCtx = ctx
			q.cancelSong = cancel

			go func(ctx context.Context, cancelFunc context.CancelFunc) {
				defer cancelFunc()
				q.CurrentlyPlaying = true
				session.UpdateGameStatus(0, song.Title)
				err := playSong(ctx, song, q.voiceConnection, q.youtubeClient)
				if err != nil {
					fmt.Println("Error playing song:", err)
				}
				q.CurrentlyPlaying = false
			}(ctx, cancel)
		}

		select {
		case <-q.skipChan:
			q.cancelSong()
		case <-q.currentSongCtx.Done():
			fmt.Println("Skipped song")
			session.UpdateGameStatus(0, "/verzoek <youtube link>")
			q.CurrentlyPlaying = false
		}
	}
	q.CurrentlyPlaying = false
}

func playSong(ctx context.Context, video *youtube.Video, voiceConnection *discordgo.VoiceConnection, client *youtube.Client) error {
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"
	options.CompressionLevel = 0
	options.PacketLoss = 10

	format := video.Formats.WithAudioChannels()
	fmt.Println("Format: ", format)
	downloadURL, err := client.GetStreamURL(video, &format[0])
	if err != nil {
		fmt.Println("Error getting stream url: ", err)
	}

	encodingSession, err := dca.EncodeFile(downloadURL, options)
	if err != nil {
		fmt.Println("Error encoding file: ", err)
	}
	defer encodingSession.Cleanup()

	done := make(chan error)
	dca.NewStream(encodingSession, voiceConnection, done)
	select {
	case err = <-done:
		if err != nil && err != io.EOF {
			fmt.Println("Error streaming: ", err)
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
