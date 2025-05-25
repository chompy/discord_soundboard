package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"text/template"

	"github.com/gorilla/websocket"
)

type PageData struct {
	GuildID      string
	ChannelID    string
	GuildName    string
	ChannelName  string
	Sounds       []string
	Categories   Categories
	ReplaceWords map[string]string
}

var httpApp *App = nil
var mainPageTmpl = template.Must(template.New("page.html.tmpl").Funcs(template.FuncMap{
	"soundNiceName": NiceName,
}).ParseFiles("client/page.html.tmpl"))
var upgrader = websocket.Upgrader{}

func buildPageDataFromRequest(r *http.Request) (PageData, error) {
	pd := PageData{GuildID: r.URL.Query().Get("guild"), ChannelID: r.URL.Query().Get("channel")}
	if pd.GuildID == "" || pd.ChannelID == "" {
		return pd, errMissingParam
	}
	guild, err := httpApp.discord.Guild(pd.GuildID)
	if err != nil {
		return pd, err
	}
	pd.GuildName = guild.Name
	channel, err := httpApp.discord.Channel(pd.ChannelID)
	if err != nil {
		return pd, err
	}
	pd.ChannelName = channel.Name
	pd.Sounds = make([]string, 0)
	for name := range httpApp.sounds {
		pd.Sounds = append(pd.Sounds, name)
	}
	pd.Categories = httpApp.Config.Categories
	pd.ReplaceWords = httpApp.Config.ReplaceWords

	sort.Strings(pd.Sounds)

	return pd, nil
}

func renderMainPage(w http.ResponseWriter, r *http.Request) {
	pd, err := buildPageDataFromRequest(r)
	if err != nil {
		httpError(w, err)
		return
	}
	mainPageTmpl.Execute(w, pd)
}

func httpJsonOutput(w http.ResponseWriter, data interface{}, statusCode int) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("An error has occured: " + err.Error()))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}

func httpServeClientJS(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "client/app.js")
}

func httpError(w http.ResponseWriter, err error) {
	out := map[string]interface{}{"status": http.StatusInternalServerError, "message": err.Error()}
	httpJsonOutput(w, out, http.StatusInternalServerError)
}

func httpPlaySound(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild")
	channelID := r.URL.Query().Get("channel")
	soundName := r.URL.Query().Get("sound")
	if guildID == "" || channelID == "" || soundName == "" {
		httpError(w, errMissingParam)
		return
	}

	vs := httpApp.VoiceSession(guildID)
	if err := vs.Play(soundName, httpApp, channelID); err != nil {
		httpError(w, err)
		return
	}

	httpJsonOutput(w, map[string]interface{}{"status": http.StatusOK}, http.StatusOK)
}

func httpPlayMultiSound(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild")
	channelID := r.URL.Query().Get("channel")
	instructions := r.URL.Query().Get("instructs")
	if guildID == "" || channelID == "" || instructions == "" {
		httpError(w, errMissingParam)
		return
	}

	vs := httpApp.VoiceSession(guildID)
	if err := vs.PlayMulti(instructions, httpApp, channelID); err != nil {
		httpError(w, err)
		return
	}

	httpJsonOutput(w, map[string]interface{}{"status": http.StatusOK}, http.StatusOK)
}

func httpStopSound(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild")
	channelID := r.URL.Query().Get("channel")
	if guildID == "" || channelID == "" {
		httpError(w, errMissingParam)
		return
	}

	vs := httpApp.VoiceSession(guildID)
	vs.Stop()

	httpJsonOutput(w, map[string]interface{}{"status": http.StatusOK}, http.StatusOK)
}

func httpReload(w http.ResponseWriter, r *http.Request) {
	if err := httpApp.loadAllSounds(); err != nil {
		httpError(w, err)
		return
	}
	httpJsonOutput(w, map[string]interface{}{"status": http.StatusOK}, http.StatusOK)
}

func httpWebSocket(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild")
	channelID := r.URL.Query().Get("channel")
	if guildID == "" || channelID == "" {
		httpError(w, errMissingParam)
		return
	}
	vs := httpApp.VoiceSession(guildID)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		httpError(w, err)
		return
	}
	defer c.Close()
	for {
		_, rawMessage, err := c.ReadMessage()
		if err != nil {
			log.Printf("> WS ERROR: %s", err.Error())
			break
		}
		messageParts := strings.Split(string(rawMessage), "|")
		switch messageParts[0] {
		case "play":
			{
				if len(messageParts) < 2 {
					log.Println("> WS ERROR (play): Malformed message.")
					break
				}
				if err := vs.Play(messageParts[1], httpApp, channelID); err != nil {
					log.Printf("> WS ERROR (play): %s", err.Error())
					break
				}
				break
			}
		case "play-multi":
			{
				if len(messageParts) < 2 {
					log.Println("> WS ERROR (play-multi): Malformed message.")
					break
				}
				if err := vs.PlayMulti(messageParts[1], httpApp, channelID); err != nil {
					log.Printf("> WS ERROR (play-multi): %s", err.Error())
					return
				}
			}
		case "stop":
			{
				vs.Stop()
				break
			}
		}

	}
}

func httpStart(app *App) error {
	httpApp = app
	http.HandleFunc("/", renderMainPage)
	http.HandleFunc("/ws", httpWebSocket)
	http.HandleFunc("/play", httpPlaySound)
	http.HandleFunc("/playm", httpPlayMultiSound)
	http.HandleFunc("/stop", httpStopSound)
	http.HandleFunc("/reload", httpReload)
	http.HandleFunc("/app.js", httpServeClientJS)
	return http.ListenAndServe(":8081", nil)
}
