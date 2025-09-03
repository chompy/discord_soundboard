package main

import (
	"database/sql"

	disgoauth "github.com/realTristan/disgoauth"
)

type App struct {
	Config      Config
	Discord     *Discord
	DiscordAuth *disgoauth.Client
}

func Run() (*App, error) {

	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if err := databaseInit(&config); err != nil {
		return nil, err
	}

	discord, err := initDiscord(&config)
	if err != nil {
		return nil, err
	}

	app := &App{Discord: discord, Config: config, DiscordAuth: initDiscordAuth(&config)}
	go RunWebServer(app)
	return app, nil
}

func (a *App) Close() error {
	return a.Discord.Close()
}

func (a *App) Database() (*sql.DB, error) {
	return databaseOpen(&a.Config)
}
