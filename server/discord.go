package main

import (
	"database/sql"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	session       *discordgo.Session
	voiceSessions []*DiscordVoiceSession
	guilds        []*discordgo.UserGuild
}

func initDiscord(config *Config) (*Discord, error) {
	log.Println("> Init Discord bot.")

	session, err := discordgo.New("Bot " + config.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	d := &Discord{session: session, voiceSessions: make([]*DiscordVoiceSession, 0), guilds: make([]*discordgo.UserGuild, 0)}
	d.session.StateEnabled = true
	d.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates | discordgo.IntentGuildMembers
	d.session.AddHandler(d.onDiscordReady)
	d.session.AddHandler(d.onDiscordVoiceStateUpdate)
	// TODO do we need a handler to know when bot is added to a server?

	if err := d.session.Open(); err != nil {
		return nil, err
	}

	if _, err := d.AvailableGuilds(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Discord) Close() error {
	return d.session.Close()
}

func (d *Discord) VoiceSession(guildID string) *DiscordVoiceSession {
	for _, vs := range d.voiceSessions {
		if vs.GuildID == guildID {
			if !vs.IsActive {
				go vs.start()
			}
			return vs
		}
	}
	vs := InitVoiceSession(guildID, d.session)
	d.voiceSessions = append(d.voiceSessions, vs)
	return vs
}

// AvailableGuilds fetches guilds that the bot is in.
func (d *Discord) AvailableGuilds() ([]*discordgo.UserGuild, error) {
	if len(d.guilds) > 0 {
		return d.guilds, nil
	}
	log.Println("> Fetch bot guild list")
	var err error
	d.guilds, err = d.session.UserGuilds(200, "", "", false)
	if err != nil {
		return nil, err
	}
	// TODO add pagination support?
	log.Printf("  - %d guild(s) found", len(d.guilds))
	return d.guilds, nil
}

func (d *Discord) UserAvailableGuilds(db *sql.DB, userId string) ([]*discordgo.UserGuild, error) {
	userGuilds, err := DatabaseGetUserGuildsByUserID(db, userId)
	if err != nil {
		return nil, err
	}
	out := make([]*discordgo.UserGuild, 0)
	for _, userGuild := range userGuilds {
		for _, botGuild := range d.guilds {
			if userGuild.GuildID == botGuild.ID {
				out = append(out, botGuild)
			}
		}
	}
	return out, nil
}

func (d *Discord) UserVoiceChannel(guildID string, userId string) (string, error) {
	vs, err := d.session.State.VoiceState(guildID, userId)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return "", nil
		}
		return "", err
	}
	return vs.ChannelID, nil
}

func (d *Discord) onDiscordReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "wiener sounds.")
}

func (d *Discord) onDiscordVoiceStateUpdate(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	if event.BeforeUpdate != nil {
		for _, vs := range d.voiceSessions {
			if vs.ChannelID == event.BeforeUpdate.ChannelID {
				if !vs.HasUsers() {
					log.Printf("> G:%s | No users in voice channel, ending session.", vs.GuildID)
					if err := vs.End(); err != nil {
						log.Printf("> G:%s | WARNING: Failed to end voice session: %s", vs.GuildID, err)
						return
					}
				}
			}
		}
	}
}
