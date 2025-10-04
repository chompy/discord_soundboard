package app

import (
	"path"

	"github.com/chompy/discord_soundboard/internal/database"
	"github.com/chompy/discord_soundboard/internal/discord"
	"github.com/chompy/discord_soundboard/internal/sound"
	"github.com/chompy/discord_soundboard/internal/web"
	"github.com/realTristan/disgoauth"
	"github.com/rs/zerolog"
)

type Client struct {
	Config Config
	logger *zerolog.Logger
}

func New() (*Client, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	logger := getLogger()
	return &Client{Config: config, logger: &logger}, nil
}

func (c *Client) Discord() (*discord.Client, error) {
	return discord.New(c.Config.DiscordBotToken, c.logger)
}

func (c *Client) DiscordAuth(redirectURL string) *disgoauth.Client {
	return discord.NewAuthClient(c.Config.DiscordAuthClientId, c.Config.DiscordAuthClientSecret, redirectURL)
}

func (c *Client) Database() (*database.Client, error) {
	return database.New(c.databasePath(), c.logger)
}

func (c *Client) Sound() *sound.Client {
	return sound.New(c.Config.StoragePath, c.logger)
}

func (c *Client) ServeHTTP() error {
	database, err := c.Database()
	if err != nil {
		return err
	}
	discord, err := c.Discord()
	if err != nil {
		return err
	}
	return web.Serve(c.Config.HTTPPort, database, discord, c.DiscordAuth(c.Config.AuthRedirectURI), c.Sound(), c.logger)
}

func (c *Client) databasePath() string {
	return path.Join(c.Config.StoragePath, "database.sqlite")
}
