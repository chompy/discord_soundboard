package main

type App struct {
	Discord *Discord
}

func Run() (*App, error) {

	if err := databaseInit(); err != nil {
		return nil, err
	}

	discord, err := NewDiscord()
	if err != nil {
		return nil, err
	}

	app := &App{Discord: discord}
	go RunWebServer(app)
	return app, nil
}

func (a *App) Close() error {
	return a.Discord.Close()
}
