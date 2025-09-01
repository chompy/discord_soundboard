package main

import (
	"io"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordVoiceSession struct {
	GuildID        string
	ChannelID      string
	IsActive       bool
	IsPlaying      bool
	discordSession *discordgo.Session
	conn           *discordgo.VoiceConnection
	buffer         *SoundReader
	lastActivity   time.Time
}

func InitVoiceSession(guild string, discordSession *discordgo.Session) *DiscordVoiceSession {
	v := &DiscordVoiceSession{
		GuildID:        guild,
		IsActive:       true,
		IsPlaying:      false,
		discordSession: discordSession,
		buffer:         nil,
		lastActivity:   time.Now(),
	}
	go v.start()
	return v
}

func (v *DiscordVoiceSession) checkConnection() error {
	if v.conn == nil {
		log.Printf("- (Re)connecting to voice session %s/%s", v.GuildID, v.ChannelID)
		var err error
		v.conn, err = v.discordSession.ChannelVoiceJoin(v.GuildID, v.ChannelID, false, true)
		if err != nil {
			v.End()
			return err
		}
	}
	return nil
}

func (v *DiscordVoiceSession) processBuffer() error {
	if err := v.checkConnection(); err != nil {
		return err
	}

	if !v.conn.Ready {
		return nil
	}

	for {
		if v.buffer == nil {
			if v.IsPlaying {
				log.Printf("- End playback in voice session %s/%s", v.GuildID, v.ChannelID)
				v.conn.Speaking(false)
				v.IsPlaying = false
			}
			break
		}
		frame, err := v.buffer.NextFrame()
		if err != nil {
			v.buffer = nil
			if err == io.EOF {
				break
			}
			return err
		}
		if !v.IsPlaying {
			log.Printf("- Begin playback in voice session %s/%s", v.GuildID, v.ChannelID)
			v.IsPlaying = true
			if err := v.conn.Speaking(true); err != nil {
				return err
			}
		}
		v.conn.OpusSend <- frame
	}

	return nil
}

func (v *DiscordVoiceSession) start() {
	log.Printf("- Begin voice session %s/%s", v.GuildID, v.ChannelID)
	for range time.Tick(time.Millisecond * 5) {
		if !v.IsActive {
			log.Printf("- End voice session %s/%s", v.GuildID, v.ChannelID)
			return
		}
		if v.ChannelID != "" {
			if err := v.processBuffer(); err != nil {
				log.Printf("- Error in voice session %s/%s: %s", v.GuildID, v.ChannelID, err)
			}
		}
	}
}

// End the voice session.
func (v *DiscordVoiceSession) End() error {
	v.IsPlaying = false
	v.buffer = nil
	var err error
	if v.conn != nil {
		err = v.conn.Disconnect()
		v.conn = nil
	}
	return err
}

// Stop play back in voice session.
func (v *DiscordVoiceSession) Stop() {
	v.buffer = nil
}

func (v *DiscordVoiceSession) Play(sound *SoundReader, channelID string) error {
	if v.ChannelID != channelID {
		v.ChannelID = channelID
		if v.conn != nil {
			v.conn.Disconnect()
			v.conn = nil
		}
	}
	//log.Printf("- Play '%s' in voice session %s/%s", sound.name, v.GuildID, v.ChannelID)
	v.buffer = sound
	v.lastActivity = time.Now()
	return nil
}

// PlayMulti plays multiple sound clips based on instructions in string.
/*func (v *DiscordVoiceSession) PlayMulti(instructionList string, app *App, channelID string) error {
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
func (v *DiscordVoiceSession) HasUsers() bool {
	members, err := v.discordSession.GuildMembers(v.GuildID, "", 100)
	if err != nil {
		log.Println("> WARNING: Failed to fetch member list: ", err)
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
