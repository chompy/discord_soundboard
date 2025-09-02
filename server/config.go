package main

import (
	"os"
)

var (
	storagePath     = os.Getenv("STORAGE_PATH")
	discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
)
