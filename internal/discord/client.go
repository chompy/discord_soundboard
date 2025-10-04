package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type Client struct {
	isOpen        bool
	session       *discordgo.Session
	voiceSessions []*VoiceSession
	guilds        []*discordgo.UserGuild
	logger        *zerolog.Logger
}

func New(token string, logger *zerolog.Logger) (*Client, error) {
	logger.Info().Msg("Init Discord bot client")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	c := &Client{isOpen: true, session: session, voiceSessions: make([]*VoiceSession, 0), guilds: make([]*discordgo.UserGuild, 0), logger: logger}
	c.session.StateEnabled = true
	c.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates | discordgo.IntentGuildMembers
	c.session.AddHandler(c.onDiscordReady)
	c.session.AddHandler(c.onDiscordVoiceStateUpdate)
	// TODO do we need a handler to know when bot is added to a server?

	if err := c.session.Open(); err != nil {
		return nil, err
	}
	if _, err := c.AvailableGuilds(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Close() error {
	c.isOpen = false
	for _, vs := range c.voiceSessions {
		vs.Close()
	}
	return c.session.Close()
}

func (c *Client) VoiceSession(guildID string) (*VoiceSession, error) {
	if !c.isOpen {
		return nil, errClientClosed
	}
	c.cleanUpVoiceSessions()
	vs := c.findVoiceSessions(guildID)
	if vs != nil {
		return vs, nil
	}
	vs = NewVoiceSession(guildID, c.session, c.logger)
	c.voiceSessions = append(c.voiceSessions, vs)
	return vs, nil
}

// AvailableGuilds fetches guilds that the bot is in.
func (c *Client) AvailableGuilds() ([]*discordgo.UserGuild, error) {
	if !c.isOpen {
		return nil, errClientClosed
	}
	if len(c.guilds) > 0 {
		return c.guilds, nil
	}

	c.logger.Info().Msg("Fetch bot guild list")

	var err error
	c.guilds, err = c.session.UserGuilds(200, "", "", false)
	if err != nil {
		return nil, err
	}
	// TODO add pagination support?

	c.logger.Info().Msgf("%d guild(s) found", len(c.guilds))

	return c.guilds, nil
}

func (c *Client) UserVoiceChannel(guildID string, userID string) (string, error) {
	vs, err := c.session.State.VoiceState(guildID, userID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return "", nil
		}
		return "", err
	}
	return vs.ChannelID, nil
}

func (c *Client) VoiceChannelHasUser(guildID string, channelID string) (bool, error) {
	members, err := c.session.GuildMembers(guildID, "", 100)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		// ignore this bot
		if m.User.ID == c.session.State.User.ID {
			continue
		}
		// check if user has active voice state
		mvs, _ := c.session.State.VoiceState(guildID, m.User.ID)
		// check if voice state is the same channel
		if mvs != nil && mvs.ChannelID == channelID {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) cleanUpVoiceSessions() {
	voiceSessions := make([]*VoiceSession, 0)
	for _, vs := range c.voiceSessions {
		if vs.IsActive {
			voiceSessions = append(voiceSessions, vs)
		}
	}
	c.voiceSessions = voiceSessions
}

func (c *Client) findVoiceSessions(guildID string) *VoiceSession {
	for _, vs := range c.voiceSessions {
		if vs.IsActive && vs.GuildID == guildID {
			return vs
		}
	}
	return nil
}

func (c *Client) onDiscordReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "wiener sounds.")
}

func (c *Client) onDiscordVoiceStateUpdate(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	if event.BeforeUpdate != nil {
		for _, vs := range c.voiceSessions {
			if vs.ChannelID == event.BeforeUpdate.ChannelID {
				hasUser, err := c.VoiceChannelHasUser(vs.GuildID, vs.ChannelID)
				if !hasUser || err != nil {
					vs.Close()
				}
			}
		}
	}
}
