package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	BotToken  string `yaml:"bot_token"`
	SoundPath string `yaml:"sound_path"`
	Timeout   int64  `yaml:"timeout"`
}

func init() {
	config.BotToken, _ = os.LookupEnv("SOUNDBOARD_BOT_TOKEN")
	config.SoundPath, _ = os.LookupEnv("SOUNDBOARD_DCA_PATH")
	timeoutStr, _ := os.LookupEnv("SOUNDBOARD_CHANNEL_TIMEOUT")
	config.Timeout = 300
	if timeoutStr != "" {
		config.Timeout, _ = strconv.ParseInt(timeoutStr, 10, 32)
	}
}

var config Config
var sounds = map[string][][]byte{}
var channelLastActivity = map[string]time.Time{}

func main() {
	if config.BotToken == "" {
		log.Fatal("> ERROR: ", errNoToken)
	}

	if err := loadSounds(); err != nil {
		log.Fatal("> ERROR: Failed to load sounds: ", err)
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + config.BotToken)
	if err != nil {
		log.Fatal("> ERROR: Failed to create Discord session: ", err)
	}

	dg.AddHandler(ready)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal("> ERROR: Failed to open Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("> Soundboard bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "MYSTIC HEROES®")
	httpDiscordSession = s
	go checkForTimeouts(s)
	go httpStart()
}

// loadSounds itterates over sounds in the sound directory and loads them in to memory
func loadSounds() error {
	dirRead, err := os.ReadDir(config.SoundPath)
	if err != nil {
		return err
	}
	for _, file := range dirRead {
		fullPath := filepath.Join(config.SoundPath, file.Name())
		if filepath.Ext(fullPath) == ".dca" {
			log.Printf("> Load sound '%s.'", fullPath)
			data, err := loadSound(fullPath)
			if err != nil {
				log.Println("> WARNING: Load sound error: ", err)
				continue
			}
			sounds[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = data
		}
	}
	return nil
}

// loadSound attempts to load an encoded sound file from disk.
func loadSound(path string) ([][]byte, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var opuslen int16

	out := make([][]byte, 0)

	for {
		// Read opus frame length from dca file.
		err := binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return nil, err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			return nil, err
		}

		// Append encoded pcm data to the buffer.
		out = append(out, InBuf)
	}

	return out, nil
}

// playSound plays the current buffer to the provided channel.
func playSound(name string, s *discordgo.Session, guildID, channelID string) (err error) {

	if sounds[name] == nil {
		return nil
	}

	log.Printf("> Play sound '%s' in voice channel '%s'.", name, channelID)

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// log last activity time
	channelLastActivity[vc.ChannelID] = time.Now()

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range sounds[name] {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	return nil
}

func checkForTimeouts(s *discordgo.Session) {
	for range time.Tick(time.Second) {
		for channelID, lastActivity := range channelLastActivity {
			timeElasped := time.Since(lastActivity)
			if timeElasped > time.Duration(config.Timeout)*time.Second {
				for _, vc := range s.VoiceConnections {
					if vc.ChannelID == channelID {
						log.Printf("> Timeout in channel '%s', leaving.", channelID)
						channelLastActivity[channelID] = time.Now().AddDate(1000, 0, 0)
						vc.Disconnect()
						break
					}
				}
			}
		}
	}
}
