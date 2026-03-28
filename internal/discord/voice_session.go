package discord

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chompy/discord_soundboard/internal/sound"
	"github.com/rs/zerolog"
)

type VoiceSession struct {
	GuildID        string
	ChannelID      string
	IsActive       bool
	IsPlaying      bool
	discordSession *discordgo.Session
	conn           *discordgo.VoiceConnection
	buffer         bytes.Buffer
	close          chan bool
	logger         *zerolog.Logger
}

func NewVoiceSession(guildID string, discordSession *discordgo.Session, logger *zerolog.Logger) *VoiceSession {
	newLogger := logger.With().Str("guildID", guildID).Logger()
	v := &VoiceSession{
		GuildID:        guildID,
		IsPlaying:      false,
		IsActive:       true,
		discordSession: discordSession,
		close:          make(chan bool),
		logger:         &newLogger,
	}

	newLogger.Info().Msgf("Begin voice session")
	voiceSessionCtx := newLogger.WithContext(context.Background())

	ticker := time.NewTicker(5 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := v.processBuffer(voiceSessionCtx); err != nil {
					newLogger.Error().Err(err).Msg("Error while processing voice session buffer")
					ticker.Stop()
					v.handleClose(voiceSessionCtx)
					return
				}

			case <-voiceSessionCtx.Done():
				v.Close()

			case <-v.close:
				ticker.Stop()
				v.handleClose(voiceSessionCtx)
				return
			}
		}
	}()

	return v
}

func (v *VoiceSession) Close() {
	v.close <- true
}

func (v *VoiceSession) checkConnection(ctx context.Context) error {
	if !v.IsActive || v.GuildID == "" || v.ChannelID == "" {
		return errNoActiveChannel
	}
	if v.conn == nil {
		logger := zerolog.Ctx(ctx)
		logger.Info().Str("channelID", v.ChannelID).Msg("Connecting to channel")
		var err error
		v.conn, err = v.discordSession.ChannelVoiceJoin(ctx, v.GuildID, v.ChannelID, false, true)
		if err != nil {
			v.Close()
			return err
		}
	}
	return nil
}

func (v *VoiceSession) processBuffer(ctx context.Context) error {
	if err := v.checkConnection(ctx); err != nil {
		return err
	}

	if v.conn.Status == discordgo.VoiceConnectionStatusReady {

		// empty buffer, no longer playing
		if v.buffer.Len() == 0 {
			if v.IsPlaying {
				logger := zerolog.Ctx(ctx)
				logger.Info().Str("channelID", v.ChannelID).Msgf("End playback in channel %s", v.ChannelID)
				v.conn.Speaking(false)
				v.IsPlaying = false
			}
			return nil
		}

		// read next frame, send to voice session
		frame, err := sound.ReadOpusFrame(&v.buffer)

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if !v.IsPlaying {
			logger := zerolog.Ctx(ctx)
			logger.Info().Str("channelID", v.ChannelID).Msgf("Begin playback in channel %s", v.ChannelID)
			v.IsPlaying = true
			if err := v.conn.Speaking(true); err != nil {
				return err
			}
		}
		v.conn.OpusSend <- frame

	}

	return nil
}

func (v *VoiceSession) handleClose(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("Close voice session")
	v.IsPlaying = false
	v.IsActive = false
	v.buffer = bytes.Buffer{}
	if v.conn != nil {
		if err := v.conn.Disconnect(ctx); err != nil {
			logger.Error().Err(err).Msg("Error while closing voice session")
		}
		v.conn = nil
	}
}

func (v *VoiceSession) Stop() {
	v.buffer = bytes.Buffer{}
}

func (v *VoiceSession) Play(sound io.Reader, channelID string) error {
	if channelID == "" {
		return errNoActiveChannel
	}
	if v.ChannelID != channelID {

		v.logger.Info().Str("channelID", channelID).Msgf("Change active channel to %s", channelID)

		if v.conn != nil {
			if err := v.conn.Disconnect(context.Background()); err != nil {
				v.logger.Warn().Str("channelID", channelID).Err(err).Msgf("Unable to close channel %s", v.ChannelID)
			}
			v.conn = nil
		}
		v.ChannelID = channelID
	}
	v.buffer = bytes.Buffer{}
	if _, err := io.Copy(&v.buffer, sound); err != nil {
		return err
	}
	return nil
}

// PlayMulti plays multiple sound clips based on instructions in string.
/*func (v *VoiceSession) PlayMulti(instructionList string, app *App, channelID string) error {
	v.buffer = make([][]byte, 0)
	if v.ChannelID != channelID {
		v.ChannelID = channelID
		if v.conn != nil {
			v.conn.Disconnect()
			v.conn = nil
		}
	}

	for _, instructionSet := range strings.Split(instructionList, ",") {
		instructionSplit := strings.Split(instructionSet, ":")
		name := strings.TrimSpace(instructionSplit[0])
		if name == "" {
			return errInvalidInstruction
		}

		sound, err := app.LoadSound(name)
		if err != nil {
			return err
		}

		start, stop := int64(0), int64(-1)
		if len(instructionSplit) > 1 {
			durSplit := strings.Split(instructionSplit[1], "-")
			start, _ = strconv.ParseInt(durSplit[0], 10, 64)
			if len(durSplit) > 1 {
				stop, _ = strconv.ParseInt(durSplit[1], 10, 64)
			}
		}
		nextBuffer := slices.Clone(sound.data)
		if start/20 <= int64(len(nextBuffer)) && stop/20 <= int64(len(nextBuffer)) {
			if start >= 0 && stop <= 0 {
				nextBuffer = nextBuffer[start/20:]
			} else if start >= 0 && stop > 0 {
				nextBuffer = nextBuffer[start/20 : stop/20]
			}
		}
		v.buffer = append(v.buffer, nextBuffer...)
	}
	log.Printf("> Play multi-sound in channel '%s'. (%s)", v.ChannelID, instructionList)
	return nil
}*/

// HasUsers checks if current voice session channel has users to broadcast to.
func (v *VoiceSession) HasUsers() bool {
	members, err := v.discordSession.GuildMembers(v.GuildID, "", 100)
	if err != nil {
		v.logger.Warn().Err(err).Msg("Failed to fetch member list")
		return false
	}
	for _, m := range members {
		// ignore this bot
		if m.User.ID == v.discordSession.State.User.ID {
			continue
		}
		// check if user has active voice state
		mvs, _ := v.discordSession.State.VoiceState(v.GuildID, m.User.ID)
		// check if voice state is the same channel as this bot
		if mvs != nil && mvs.ChannelID == v.ChannelID {
			return true
		}
	}
	return false
}
