package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type App struct {
	Config        Config
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
	soundPath, _ := os.LookupEnv("SOUNDBOARD_SOUND_PATH")
	if soundPath == "" {
		soundPath = "./sounds"
	}
	// load config data from yaml file
	var err error
	a.Config, err = LoadConfig(soundPath)
	if err != nil {
		return err
	}

	a.voiceSessions = make([]*VoiceSession, 0)

	// load sounds
	if err := a.loadAllSounds(); err != nil {
		return err
	}

	// Create a new Discord session using the provided bot token.
	a.discord, err = discordgo.New("Bot " + botToken)
	if err != nil {
		return err
	}
	a.discord.StateEnabled = true
	a.discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates | discordgo.IntentGuildMembers
	a.discord.AddHandler(a.onDiscordReady)
	a.discord.AddHandler(a.onDiscordVoiceStateUpdate)
	return a.discord.Open()
}

func (a *App) Close() error {
	return a.discord.Close()
}

func (a *App) VoiceSession(guildID string) *VoiceSession {
	for _, vs := range a.voiceSessions {
		if vs.GuildID == guildID {
			return vs
		}
	}
	vs := &VoiceSession{GuildID: guildID, buffer: make([][]byte, 0), lastActivity: time.Now()}
	a.voiceSessions = append(a.voiceSessions, vs)
	return vs
}

func (a *App) onDiscordReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, fmt.Sprintf("%d wiener loving sounds", len(a.sounds)))
	go httpStart(a)
	go a.appLoop()
}

func (a *App) onDiscordVoiceStateUpdate(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {

	// only want to track user's leaving
	if event.BeforeUpdate == nil {
		return
	}

	// determine if event is in channel where bot resides
	for _, vs := range a.voiceSessions {
		if vs.ChannelID == event.BeforeUpdate.ChannelID {
			if !vs.HasUsers(s) {
				log.Printf("> Leaving channel '%s', no users left.", vs.ChannelID)
				if err := vs.End(); err != nil {
					log.Println("> WARNING: Failed to end voice session: ", err)
					return
				}
			}
		}
	}

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
