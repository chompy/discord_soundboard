package main

import (
	"github.com/chompy/discord_soundboard/internal/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		panic(err)
	}
	if err := app.ServeHTTP(); err != nil {
		panic(err)
	}
}
