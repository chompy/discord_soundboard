package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chompy/discord_soundboard/internal/app"
)

func main() {

	app := &app.App{}
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("> Soundboard bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	if err := app.Close(); err != nil {
		log.Fatal(err)
	}
}
