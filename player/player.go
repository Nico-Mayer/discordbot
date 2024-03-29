package player

import (
	"github.com/ClintonCollins/dca"
	"github.com/bwmarrin/discordgo"
)

type PlayerStatus int32

const (
	Resting PlayerStatus = 0
	Playing PlayerStatus = 1
	Paused  PlayerStatus = 2
	Err     PlayerStatus = 3
)

type Player struct {
	Session       *discordgo.Session
	QueueList     []string
	PlayerStatus  PlayerStatus
	voiceConn     *discordgo.VoiceConnection
	queue         chan *Song
	skipInterrupt chan bool
	currentStream *dca.StreamingSession
	options       *dca.EncodeOptions
}

func NewPlayer(s *discordgo.Session) *Player {
	return &Player{
		Session:       s,
		queue:         make(chan *Song, 100),
		skipInterrupt: make(chan bool, 1),
		PlayerStatus:  Resting,
		options: &dca.EncodeOptions{
			Volume:           100,
			Channels:         2,
			FrameRate:        48000,
			FrameDuration:    20,
			Bitrate:          64,
			Application:      dca.AudioApplicationLowDelay,
			CompressionLevel: 7,
			PacketLoss:       3,
			BufferedFrames:   200,
			VBR:              true,
			StartTime:        0,
			VolumeFloat:      1.0,
			RawOutput:        true,
		},
	}
}

func (p *Player) JoinChannel(vs *discordgo.VoiceState) error {
	if p.voiceConn != nil {
		return nil
	}

	if p.voiceConn == nil {
		conn, err := p.Session.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
		if err != nil {
			return err
		}
		p.voiceConn = conn
	}
	return nil
}
