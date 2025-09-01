package main

import (
	"os"
)

var (
	appName         = "Chompy's Sound Board"
	storagePath     = os.Getenv("STORAGE_PATH")
	discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
)
