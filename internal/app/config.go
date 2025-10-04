package app

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const envVarPrefix = "SOUNDBOARD_"

var (
	envPathes = []string{".env", ".env.local"}
)

type Config struct {
	HTTPPort                int
	AuthRedirectURI         string
	StoragePath             string
	SessionSecret           string
	DiscordBotToken         string
	DiscordAuthClientId     string
	DiscordAuthClientSecret string
}

func LoadConfig() (Config, error) {
	for _, envPath := range envPathes {
		godotenv.Load(envPath)
	}

	config := Config{}
	config.HTTPPort, _ = strconv.Atoi(os.Getenv(envVarPrefix + "HTTP_PORT"))
	if config.HTTPPort == 0 {
		config.HTTPPort = 8081
	}
	config.AuthRedirectURI = os.Getenv(envVarPrefix + "REDIRECT_URI")
	if config.AuthRedirectURI == "" {
		config.AuthRedirectURI = fmt.Sprintf("http://localhost:%d/redirect", config.HTTPPort)
	}
	config.StoragePath = os.Getenv(envVarPrefix + "STORAGE_PATH")
	if config.StoragePath == "" {
		config.StoragePath = "storage"
	}
	config.SessionSecret = os.Getenv(envVarPrefix + "SESSION_SECRET")
	if config.SessionSecret == "" {
		config.SessionSecret = "--sound!fb7ccb7f8015e01fb23bbe08084f913748cdc2c5193d1f0cdb1a82903e1bae67"
	}
	config.DiscordBotToken = os.Getenv(envVarPrefix + "DISCORD_BOT_TOKEN")
	if config.DiscordBotToken == "" {
		return config, errMissingBotToken
	}
	config.DiscordAuthClientId = os.Getenv(envVarPrefix + "DISCORD_OAUTH_CLIENT_ID")
	config.DiscordAuthClientSecret = os.Getenv(envVarPrefix + "DISCORD_OAUTH_CLIENT_SECRET")
	if config.DiscordAuthClientId == "" || config.DiscordAuthClientSecret == "" {
		return config, errMissingAuthConfig
	}

	return config, nil
}
