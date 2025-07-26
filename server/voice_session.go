package main

import (
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type VoiceSession struct {
	GuildID      string
	ChannelID    string
	conn         *discordgo.VoiceConnection
	buffer       [][]byte
	lastActivity time.Time
	isSpeaking   bool
}

// Process is the main processing loop.
func (v *VoiceSession) Process(app *App) error {

	// sound data in buffer
	if len(v.buffer) > 0 {
		if v.conn == nil {
			var err error
			v.conn, err = app.discord.ChannelVoiceJoin(v.GuildID, v.ChannelID, false, true)
			if err != nil {
				v.End()
				return err
			}
		}

		if !v.isSpeaking {
			v.isSpeaking = true
			if err := v.conn.Speaking(true); err != nil {
				return err
			}
		}
		v.lastActivity = time.Now()

		var buff []byte
		buff, v.buffer = v.buffer[0], v.buffer[1:len(v.buffer)]
		v.conn.OpusSend <- buff

	} else if v.conn != nil {
		if v.isSpeaking {
			v.conn.Speaking(false)
			v.isSpeaking = false
		}

	}

	return nil
}

// End disconnects from Discord, ending the session.
func (v *VoiceSession) End() error {
	v.isSpeaking = false
	v.buffer = make([][]byte, 0)
	if v.conn == nil {
		return nil
	}
	err := v.conn.Disconnect()
	v.conn = nil
	return err
}

// Stop stops all sounds in current Discord channel.
func (v *VoiceSession) Stop() {
	log.Printf("> Stop playback in channel '%s'.", v.ChannelID)
	v.buffer = make([][]byte, 0)
}

// Play streams sound of given name to given Discord channel.
func (v *VoiceSession) Play(name string, app *App, channelID string) error {
	sound, err := app.LoadSound(name)
	if err != nil {
		return err
	}
	return v.PlayFromData(sound, app, channelID)
}

// PlayFromData streams given sound data to given Discord channel.
func (v *VoiceSession) PlayFromData(sound *SoundData, app *App, channelID string) error {
	if sound == nil || sound.data == nil || len(sound.data) == 0 {
		return errInvalidSound
	}
	if v.ChannelID != channelID {
		v.ChannelID = channelID
		if v.conn != nil {
			v.conn.Disconnect()
			v.conn = nil
		}
	}
	log.Printf("> Play '%s' in channel '%s'.", sound.name, v.ChannelID)
	v.buffer = slices.Clone(sound.data)
	return nil
}

// PlayMulti plays multiple sound clips based on instructions in string.
func (v *VoiceSession) PlayMulti(instructionList string, app *App, channelID string) error {
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
}

func (v *VoiceSession) IsPlaying() bool {
	return v.isSpeaking
}

// HasUsers checks if current voice session channel has users to broadcast to.
func (v *VoiceSession) HasUsers(s *discordgo.Session) bool {
	members, err := s.GuildMembers(v.GuildID, "", 100)
	if err != nil {
		log.Println("> WARNING: Failed to fetch member list: ", err)
		return false
	}
	for _, m := range members {
		// ignore this bot
		if m.User.ID == s.State.User.ID {
			continue
		}
		// check if user has active voice state
		mvs, _ := s.State.VoiceState(v.GuildID, m.User.ID)
		// check if voice state is the same channel as this bot
		if mvs != nil && mvs.ChannelID == v.ChannelID {
			return true
		}
	}
	return false
}
