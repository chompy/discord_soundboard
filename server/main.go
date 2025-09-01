package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	log.Println("> Start app.")

	app, err := Run()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	// Wait here until CTRL-C or other term signal is received.
	log.Println("> Soundboard bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
