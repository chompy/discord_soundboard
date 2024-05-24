package app

import (
	"log"
	"math/rand"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

const randomSound = "yoshi-tongue"

type VoiceSession struct {
	GuildID             string
	ChannelID           string
	conn                *discordgo.VoiceConnection
	buffer              [][]byte
	lastActivity        time.Time
	isSpeaking          bool
	nextSilenceInterupt time.Time
}

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
		timeElasped := time.Since(v.lastActivity)
		if timeElasped > time.Duration(app.BotTimeout)*time.Second {
			log.Printf("> Timeout in channel '%s.'", v.ChannelID)
			return v.End()
		}
		if time.Now().After(v.nextSilenceInterupt) && app.sounds[randomSound] != nil {
			log.Println("> Interupt silence!")
			return v.Play(randomSound, app)
		}
	}

	return nil
}

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

func (v *VoiceSession) Stop() {
	log.Printf("> Stop playback in channel '%s'.", v.ChannelID)
	v.buffer = make([][]byte, 0)
}

func (v *VoiceSession) Play(name string, app *App) error {
	if app.sounds[name] == nil {
		return errSoundNotFound
	}
	log.Printf("> Play '%s' in channel '%s'.", name, v.ChannelID)
	v.nextSilenceInterupt = time.Now().Add(time.Second * time.Duration((rand.Int63n(3540) + 60)))
	v.buffer = slices.Clone(app.sounds[name])
	return nil
}

func (v *VoiceSession) IsPlaying() bool {
	return v.isSpeaking
}
