package app

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type App struct {
	BotTimeout    int64
	SoundPath     string
	sounds        map[string][][]byte
	discord       *discordgo.Session
	voiceSessions []*VoiceSession
}

func (a *App) Start() error {
	// load config data from env vars
	botToken, _ := os.LookupEnv("SOUNDBOARD_BOT_TOKEN")
	if botToken == "" {
		return errNoToken
	}
	a.SoundPath, _ = os.LookupEnv("SOUNDBOARD_SOUND_PATH")
	if a.SoundPath == "" {
		a.SoundPath = "./sounds"
	}
	timeoutStr, _ := os.LookupEnv("SOUNDBOARD_CHANNEL_TIMEOUT")
	a.BotTimeout = 300
	if timeoutStr != "" {
		a.BotTimeout, _ = strconv.ParseInt(timeoutStr, 10, 32)
	}

	a.voiceSessions = make([]*VoiceSession, 0)

	// load sounds
	if err := a.loadAllSounds(); err != nil {
		return err
	}

	// Create a new Discord session using the provided bot token.
	var err error
	a.discord, err = discordgo.New("Bot " + botToken)
	if err != nil {
		return err
	}
	a.discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
	a.discord.AddHandler(a.onDiscordReady)
	return a.discord.Open()
}

func (a *App) Close() error {
	return a.discord.Close()
}

func (a *App) VoiceSession(guildID string, channelID string) *VoiceSession {
	for _, vs := range a.voiceSessions {
		if vs.GuildID == guildID && vs.ChannelID == channelID {
			return vs
		}
	}
	vs := &VoiceSession{GuildID: guildID, ChannelID: channelID, buffer: make([][]byte, 0), lastActivity: time.Now()}
	a.voiceSessions = append(a.voiceSessions, vs)
	return vs
}

func (a *App) onDiscordReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, fmt.Sprintf("%d weiner loving sounds", len(a.sounds)))
	go httpStart(a)
	go a.appLoop()
}

func (a *App) appLoop() {
	for {
		hasPlaying := false
		for _, vs := range a.voiceSessions {
			if vs != nil {
				vs.Process(a)
				if vs.IsPlaying() {
					hasPlaying = true
				}
			}
		}
		if !hasPlaying {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
