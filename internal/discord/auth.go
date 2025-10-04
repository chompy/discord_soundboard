package discord

import (
	"github.com/realTristan/disgoauth"
)

func NewAuthClient(clientId string, clientSecret string, redirectUrl string) *disgoauth.Client {
	return disgoauth.Init(&disgoauth.Client{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURI:  redirectUrl,
		Scopes:       []string{disgoauth.ScopeIdentify, disgoauth.ScopeGuilds},
	})
}
