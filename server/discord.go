package main

import (
	"database/sql"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	session       *discordgo.Session
	voiceSessions []*DiscordVoiceSession
	Guilds        []*discordgo.UserGuild
}

func NewDiscord() (*Discord, error) {
	log.Println("> Init Discord bot.")
	if discordBotToken == "" {
		return nil, errNoToken
	}

	session, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		return nil, err
	}

	d := &Discord{session: session, voiceSessions: make([]*DiscordVoiceSession, 0), Guilds: make([]*discordgo.UserGuild, 0)}
	d.session.StateEnabled = true
	d.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates | discordgo.IntentGuildMembers
	d.session.AddHandler(d.onDiscordReady)
	d.session.AddHandler(d.onDiscordVoiceStateUpdate)

	log.Println("  - Open session")
	if err := d.session.Open(); err != nil {
		return nil, err
	}

	if err := d.fetchGuilds(); err != nil {
		return nil, err
	}

	log.Println("  - Done")
	return d, nil
}

func (d *Discord) Close() error {
	return d.session.Close()
}

func (d *Discord) VoiceSession(guildID string) *DiscordVoiceSession {
	for _, vs := range d.voiceSessions {
		if vs.GuildID == guildID {
			return vs
		}
	}
	vs := InitVoiceSession(guildID, d.session)
	d.voiceSessions = append(d.voiceSessions, vs)
	return vs
}

func (d *Discord) UserAvailableGuilds(db *sql.DB, userId string) ([]*discordgo.UserGuild, error) {
	userGuilds, err := fetchUserGuildsByUserID(db, userId)
	if err != nil {
		return nil, err
	}
	out := make([]*discordgo.UserGuild, 0)
	for _, userGuild := range userGuilds {
		for _, botGuild := range d.Guilds {
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

func (d *Discord) fetchGuilds() error {
	log.Println("  - Fetch bot guild list")
	guilds, err := d.session.UserGuilds(200, "", "", false)
	if err != nil {
		return err
	}
	// TODO: paginate
	d.Guilds = append(d.Guilds, guilds...)
	log.Printf("  - %d guild(s) found", len(d.Guilds))
	return nil
}

func (d *Discord) onDiscordReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "wiener sounds.")
}

func (d *Discord) onDiscordVoiceStateUpdate(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	if event.BeforeUpdate != nil {
		for _, vs := range d.voiceSessions {
			if vs.ChannelID == event.BeforeUpdate.ChannelID {
				if !vs.HasUsers() {
					if err := vs.End(); err != nil {
						log.Printf("> WARNING: Failed to end voice session %s/%s: %s", vs.GuildID, vs.ChannelID, err)
						return
					}
				}
			}
		}
	}
}
