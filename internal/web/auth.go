package web

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chompy/discord_soundboard/internal/database"
	disgoauth "github.com/realTristan/disgoauth"
)

const sessionCookieName = "_dsb_session"

func generateSessionToken(user *database.User, sessionSecret string) string {
	hashSalt := fmt.Sprintf("%s-%s-%d-%d", sessionSecret, user.ID, user.Created.UnixMicro(), time.Now().UnixMicro())
	hash := sha512.New()
	hash.Write([]byte(hashSalt))
	data := fmt.Sprintf("%s|%x", user.ID, hash.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func fetchUserIDFromSessionToken(token string) (string, error) {
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	decodedToken := string(tokenBytes[:])
	userId := strings.Split(decodedToken, "|")[0]
	return userId, nil
}

func checkUser(databaseClient *database.Client, r *http.Request) (database.User, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return database.User{}, errNotAuthenticated
	}

	userId, err := fetchUserIDFromSessionToken(c.Value)
	if err != nil {
		return database.User{}, errNotAuthenticated
	}

	user, err := databaseClient.FetchUserByID(userId)
	if err != nil {
		return user, errNotAuthenticated
	}

	if user.SessionToken != c.Value {
		return database.User{}, errNotAuthenticated
	}

	return user, nil
}

func checkUserGuild(databaseClient *database.Client, userID string, guildID string) error {
	hasGuild, err := databaseClient.UserHasGuild(userID, guildID)
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
		return nil, errInvalidAuth
	}
	return resp.Body, nil
}

func httpLogin(rctx *RequestContext) (string, error, error) {
	rctx.Services.Auth.RedirectHandler(rctx.RW, rctx.Req, "")
	return "", nil, nil
}

func httpLoginRedirect(rctx *RequestContext) (string, error, error) {
	code := rctx.Req.URL.Query().Get("code")

	if rctx.Req.URL.Query().Get("error") != "" || code == "" {
		return "", errInvalidAuth, errInvalidAuth
	}

	accessToken, _ := rctx.Services.Auth.GetOnlyAccessToken(code)
	userData, _ := disgoauth.GetUserData(accessToken)

	userID, userIDOK := userData["id"].(string)
	userName, userNameOk := userData["username"].(string)
	if !userIDOK || !userNameOk || userID == "" {
		return "", errInvalidAuth, errInvalidAuth
	}

	rctx.Logger.Info().Str("userID", userID).Msgf("User #%s (%s) has logged in", userID, userName)

	user, err := rctx.Services.Database.FetchUserByID(userID)
	if err != nil {
		return "", err, errDatabaseRead
	}

	user.ID = userID
	user.Name = userName
	user.SessionToken = generateSessionToken(&user, "TODOSESSIONSECRET!!!!@#@") // TODO
	if err := rctx.Services.Database.Save(&user); err != nil {
		return "", err, errDatabaseWrite
	}

	guildsReader, err := fetchUserGuilds(accessToken)
	if err != nil {
		return "", err, errInvalidAuth
	}

	if err := rctx.Services.Database.ImportUserGuilds(guildsReader, user.ID); err != nil {
		return "", err, errDatabaseWrite
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
	http.SetCookie(rctx.RW, &cookie)
	return "/", nil, nil
}
