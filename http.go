package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"text/template"

	"github.com/bwmarrin/discordgo"
)

type PageData struct {
	GuildID     string
	ChannelID   string
	GuildName   string
	ChannelName string
	Sounds      []string
}

var httpDiscordSession *discordgo.Session = nil
var mainPageTmpl = template.Must(template.ParseFiles("assets/page.html.tmpl"))

func buildPageDataFromRequest(r *http.Request) (PageData, error) {
	pd := PageData{GuildID: r.URL.Query().Get("guild"), ChannelID: r.URL.Query().Get("channel")}
	if pd.GuildID == "" || pd.ChannelID == "" {
		return pd, errMissingParam
	}
	guild, err := httpDiscordSession.Guild(pd.GuildID)
	if err != nil {
		return pd, err
	}
	pd.GuildName = guild.Name
	channel, err := httpDiscordSession.Channel(pd.ChannelID)
	if err != nil {
		return pd, err
	}
	pd.ChannelName = channel.Name
	pd.Sounds = make([]string, 0)
	for name := range sounds {
		pd.Sounds = append(pd.Sounds, name)
	}
	sort.Strings(pd.Sounds)

	return pd, nil
}

func renderMainPage(w http.ResponseWriter, r *http.Request) {
	pd, err := buildPageDataFromRequest(r)
	if err != nil {
		httpError(w, r, err)
		return
	}
	mainPageTmpl.Execute(w, pd)
}

func httpJsonOutput(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
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

func httpError(w http.ResponseWriter, r *http.Request, err error) {
	out := map[string]interface{}{"status": http.StatusInternalServerError, "message": err.Error()}
	httpJsonOutput(w, r, out, http.StatusInternalServerError)
}

func httpPlaySound(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Query().Get("guild")
	channelID := r.URL.Query().Get("channel")
	soundName := r.URL.Query().Get("sound")
	if guildID == "" || channelID == "" || soundName == "" {
		httpError(w, r, errMissingParam)
		return
	}
	if err := playSound(soundName, httpDiscordSession, guildID, channelID); err != nil {
		httpError(w, r, err)
		return
	}
	httpJsonOutput(w, r, map[string]interface{}{"status": http.StatusOK}, http.StatusOK)
}

func httpStart() error {
	http.HandleFunc("/", renderMainPage)
	http.HandleFunc("/play", httpPlaySound)
	return http.ListenAndServe(":8081", nil)
}
