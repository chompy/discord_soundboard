package main

import (
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	disgoauth "github.com/realTristan/disgoauth"
)

const sessionCookieName = "_dsb_session"

func initDiscordAuth(config *Config) *disgoauth.Client {
	return disgoauth.Init(&disgoauth.Client{
		ClientID:     config.DiscordAuthClientId,
		ClientSecret: config.DiscordAuthClientSecret,
		RedirectURI:  os.Getenv("SOUNDBOARD_DISCORD_OAUTH_REDIRECT_URL"), // TODO
		Scopes:       []string{disgoauth.ScopeIdentify, disgoauth.ScopeGuilds},
	})
}

func generateSessionToken(config *Config, user *User) string {
	hashSalt := fmt.Sprintf("%s-%s-%d-%d", config.SessionSecret, user.ID, user.Created.UnixMicro(), time.Now().UnixMicro())
	hash := sha512.New()
	hash.Write([]byte(hashSalt))
	data := fmt.Sprintf("%s|%x", user.ID, hash.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func fetchUserIdFromSessionToken(token string) (string, error) {
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	decodedToken := string(tokenBytes[:])
	userId := strings.Split(decodedToken, "|")[0]
	return userId, nil
}

func checkUser(db *sql.DB, r *http.Request) (User, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return User{}, errNotAuthenticated
	}

	userId, err := fetchUserIdFromSessionToken(c.Value)
	if err != nil {
		return User{}, errNotAuthenticated
	}

	user, err := DatabaseGetUserByID(db, userId)
	if err != nil {
		return user, errNotAuthenticated
	}

	if user.SessionToken != c.Value {
		return User{}, errNotAuthenticated
	}

	return user, nil
}

func checkUserGuild(db *sql.DB, userID string, guildID string) error {
	hasGuild, err := DatabaseUserHasGuild(db, userID, guildID)
	if err != nil {
		return err
	}
	if !hasGuild {
		return errNotAuthorized
	}
	return nil
}

func fetchUserGuilds(accessToken string) (io.Reader, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", accessToken)
	resp, err := disgoauth.RequestClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		o, _ := io.ReadAll(resp.Body)
		log.Println(string(o))
		return nil, errInvalidAuth
	}
	return resp.Body, nil
}

func httpLogin(app *App, w http.ResponseWriter, r *http.Request) error {
	app.DiscordAuth.RedirectHandler(w, r, "")
	return nil
}

func httpLoginRedirect(app *App, w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")

	if r.URL.Query().Get("error") != "" || code == "" {
		return errInvalidAuth
	}

	accessToken, _ := app.DiscordAuth.GetOnlyAccessToken(code)
	userData, _ := disgoauth.GetUserData(accessToken)

	db, err := app.Database()
	if err != nil {
		return errDatabaseOpen
	}
	defer db.Close()

	userId, userIdOk := userData["id"].(string)
	userName, userNameOk := userData["username"].(string)
	if !userIdOk || !userNameOk || userId == "" {
		return errInvalidAuth
	}

	log.Printf("- User #%s (%s) has logged in", userId, userData["username"])

	user, err := DatabaseGetUserByID(db, userId)
	if err != nil {
		return errDatabaseRead
	}

	user.ID = userId
	user.Name = userName
	user.SessionToken = generateSessionToken(&app.Config, &user)
	if err := user.Save(db); err != nil {
		return errDatabaseWrite
	}

	guildsReader, err := fetchUserGuilds(accessToken)
	if err != nil {
		return errInvalidAuth
	}
	if err := databaseImportUserGuilds(db, guildsReader, user.ID); err != nil {
		return errDatabaseWrite
	}

	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    user.SessionToken,
		Path:     "/",
		MaxAge:   604800, // one week
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}
