module github.com/chompy/discord_soundboard

go 1.25.1

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/realTristan/disgoauth v1.0.2
	github.com/rs/zerolog v1.34.0
	gopkg.in/hraban/opus.v2 v2.0.0-20230925203106-0188a62cb302
)

require (
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/bwmarrin/discordgo => github.com/yeongaori/discordgo-fork v0.0.0-20260308044327-f9e3cff6c311
