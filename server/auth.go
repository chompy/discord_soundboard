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

var (
	sessionSecret                   = []byte("--sound!fb7ccb7f8015e01fb23bbe08084f913748cdc2c5193d1f0cdb1a82903e1bae67")
	discordAuth   *disgoauth.Client = disgoauth.Init(&disgoauth.Client{
		ClientID:     os.Getenv("DISCORD_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_OAUTH_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("DISCORD_OAUTH_REDIRECT_URL"),
		Scopes:       []string{disgoauth.ScopeIdentify, disgoauth.ScopeGuilds},
	})
)

func generateSessionToken(user *User) string {
	hashSalt := fmt.Sprintf("%s-%s-%d-%d", sessionSecret, user.ID, user.Created.UnixMicro(), time.Now().UnixMicro())
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

	user, err := databaseFetchUserByID(db, userId)
	if err != nil {
		return user, errNotAuthenticated
	}

	if user.SessionToken != c.Value {
		return User{}, errNotAuthenticated
	}

	return user, nil
}

func checkUserGuild(db *sql.DB, userID string, guildID string) error {
	hasGuild, err := userHasGuild(db, userID, guildID)
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

func httpLogin(w http.ResponseWriter, r *http.Request) {
	discordAuth.RedirectHandler(w, r, "")
}

func httpLoginRedirect(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	if r.URL.Query().Get("error") != "" || code == "" {
		httpApiError(w, errInvalidAuth)
		return
	}

	accessToken, _ := discordAuth.GetOnlyAccessToken(code)
	userData, _ := disgoauth.GetUserData(accessToken)

	db, err := databaseOpen()
	if err != nil {
		httpApiError(w, err)
		return
	}
	defer db.Close()

	userId, userIdOk := userData["id"].(string)
	userName, userNameOk := userData["username"].(string)
	if !userIdOk || !userNameOk || userId == "" {
		httpApiError(w, errInvalidAuth)
		return
	}

	log.Printf("- User #%s (%s) has logged in", userId, userData["username"])

	user, err := databaseFetchUserByID(db, userId)
	if err != nil {
		httpApiError(w, err)
		return
	}

	user.ID = userId
	user.Name = userName
	user.SessionToken = generateSessionToken(&user)
	if err := user.Save(db); err != nil {
		httpApiError(w, err)
		return
	}

	guildsReader, err := fetchUserGuilds(accessToken)
	if err != nil {
		log.Println(err)
		httpApiError(w, errInvalidAuth)
		return
	}
	if err := importUserGuilds(db, guildsReader, user.ID); err != nil {
		httpApiError(w, err)
		return
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
}
