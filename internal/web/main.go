package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/chompy/discord_soundboard/internal/database"
	"github.com/chompy/discord_soundboard/internal/discord"
	"github.com/chompy/discord_soundboard/internal/sound"
	disgoauth "github.com/realTristan/disgoauth"
	"github.com/rs/zerolog"
)

type Services struct {
	Discord  *discord.Client
	Auth     *disgoauth.Client
	Database *database.Client
	Sound    *sound.Client
	Logger   *zerolog.Logger
}

type RequestContext struct {
	ID       int
	Req      *http.Request
	RW       http.ResponseWriter
	User     database.User
	Services *Services
	Logger   *zerolog.Logger
}

type ApiRequest struct {
	ID       int
	R        *http.Request
	W        http.ResponseWriter
	User     database.User
	Discord  *discord.Client
	Database *database.Client
}

var requestIDCounter = 0

func httpApiJsonWrite(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	jsonData, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": "` + err.Error() + `}`))
		return
	}
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}

func httpApiJsonRead(r *http.Request, data any) error {
	rawJson, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(rawJson, &data)
}

func httpApiError(rctx *RequestContext, err error, userErr error) {
	if userErr == nil {
		userErr = errUnknown
	}
	statusCode := errHttpStatusCodeMap[userErr]
	if statusCode == 0 {
		statusCode = 500
	}
	rctx.Logger.Warn().Err(err).Msgf("%d | %d | %s %s -- %s", rctx.ID, statusCode, rctx.Req.Method, rctx.Req.URL.Path, err.Error())
	httpApiJsonWrite(rctx.RW, map[string]any{"success": false, "error": userErr.Error(), "isLoggedIn": rctx.User.ID != ""}, statusCode)
}

func handleHttpApi(services *Services, callback func(rctx *RequestContext) (any, error, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestIDCounter++

		logger := services.Logger.With().Int("requestID", requestIDCounter).Logger()
		rctx := RequestContext{
			ID:       requestIDCounter,
			Req:      r,
			RW:       w,
			Services: services,
			Logger:   &logger,
		}

		var err error
		rctx.User, err = checkUser(rctx.Services.Database, r)
		if err != nil {
			httpApiError(&rctx, err, errNotAuthenticated)
			return
		}

		resp, err, userErr := callback(&rctx)
		if err != nil {
			httpApiError(&rctx, err, userErr)
			return
		}

		if resp == nil {
			resp = map[string]any{"success": true}
		}
		httpApiJsonWrite(w, resp, http.StatusOK)

		logger.Info().Msgf("%d | %d | %s %s", rctx.ID, http.StatusOK, r.Method, r.URL.Path)
	}
}

func httpApiMe(rctx *RequestContext) (any, error, error) {
	return map[string]any{"success": true, "user": rctx.User}, nil, nil
}

func httpApiListGuilds(rctx *RequestContext) (any, error, error) {
	botGuilds, err := rctx.Services.Discord.AvailableGuilds()
	if err != nil {
		return nil, err, errDiscordApi
	}
	userGuilds, err := rctx.Services.Database.FetchUserGuildsByUserID(rctx.User.ID)
	if err != nil {
		return nil, err, errDatabaseRead
	}
	userBotGuilds := make([]*discordgo.UserGuild, 0)
	for _, botGuild := range botGuilds {
		for _, userGuild := range userGuilds {
			if botGuild.ID == userGuild.GuildID {
				userBotGuilds = append(userBotGuilds, botGuild)
			}
		}
	}
	return map[string]any{"success": true, "guilds": userBotGuilds}, nil, nil
}

func httpApiListGuildCategories(rctx *RequestContext) (any, error, error) {
	guildId := rctx.Req.URL.Query().Get("guild")
	if guildId == "" {
		return nil, errMissingParam, errMissingParam
	}
	if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, guildId); err != nil {
		return nil, err, errNotAuthorized
	}

	categories, err := rctx.Services.Database.FetchCategoriesByGuildID(guildId)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	return map[string]any{"success": true, "categories": categories}, nil, nil
}

func httpApiListGuildSounds(rctx *RequestContext) (any, error, error) {
	guildID := rctx.Req.URL.Query().Get("guild")
	if guildID == "" {
		return nil, errMissingParam, errMissingParam
	}

	if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, guildID); err != nil {
		return nil, err, errNotAuthorized
	}

	sounds, err := rctx.Services.Database.FetchSoundsByGuildID(guildID)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	return map[string]any{"success": true, "sounds": sounds}, nil, nil
}

func httpApiListGuildCategoriesAndSounds(rctx *RequestContext) (any, error, error) {
	guildID := rctx.Req.URL.Query().Get("guild")
	if guildID == "" {
		return nil, errMissingParam, errMissingParam
	}

	if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, guildID); err != nil {
		return nil, err, errNotAuthorized
	}

	categories, err := rctx.Services.Database.FetchCategoriesByGuildID(guildID)
	if err != nil {
		return nil, err, errDatabaseRead
	}
	sounds, err := rctx.Services.Database.FetchSoundsByGuildID(guildID)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	return map[string]any{"success": true, "categories": categories, "sounds": sounds}, nil, nil

}

type httpDeleteCategory struct {
	ID int64 `json:"id"`
}

type httpCreateCategory struct {
	Name    string `json:"name"`
	GuildID string `json:"guildId"`
	Sort    int    `json:"sort"`
}

type httpUpdateCategory struct {
	httpDeleteCategory
	httpCreateCategory
}

func httpApiModCategory(rctx *RequestContext) (any, error, error) {
	switch rctx.Req.Method {
	case http.MethodPost:
		{
			params := httpCreateCategory{}
			if err := httpApiJsonRead(rctx.Req, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.Name == "" || params.GuildID == "" {
				return nil, errMissingParam, errMissingParam
			}

			if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, params.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			category := database.Category{Name: params.Name, GuildID: params.GuildID, Sort: params.Sort}
			if err := rctx.Services.Database.Save(&category); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "category": category}, nil, nil
		}
	case http.MethodPut:
		{
			params := httpUpdateCategory{}
			if err := httpApiJsonRead(rctx.Req, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.ID == 0 || params.Name == "" || params.GuildID == "" {
				return nil, errMissingParam, errMissingParam
			}

			if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, params.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			category, err := rctx.Services.Database.FetchCategoryByID(params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			category.Name = params.Name
			category.GuildID = params.GuildID
			category.Sort = params.Sort
			if err := rctx.Services.Database.Save(&category); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "category": category}, nil, nil
		}
	case http.MethodDelete:
		{
			params := httpDeleteCategory{}
			if err := httpApiJsonRead(rctx.Req, &params); err != nil {
				return nil, err, errInvalidParam
			}

			category, err := rctx.Services.Database.FetchCategoryByID(params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			if err := rctx.Services.Database.Delete(&category); err != nil {
				return nil, err, errDatabaseWrite
			}

			return nil, nil, nil
		}
	}
	return nil, errInvalidMethod, errInvalidMethod
}

type httpSortCategory struct {
	GuildID string  `json:"guildId"`
	IDs     []int64 `json:"ids"`
}

func httpApiSortCategory(rctx *RequestContext) (any, error, error) {
	params := httpSortCategory{}
	if err := httpApiJsonRead(rctx.Req, &params); err != nil {
		return nil, err, errInvalidParam
	}

	if err := rctx.Services.Database.SortCategories(params.GuildID, params.IDs...); err != nil {
		return nil, err, errDatabaseWrite
	}

	return nil, nil, nil
}

type httpCreateSound struct {
	Name       string `json:"name"`
	Hash       string `json:"hash"`
	CategoryID int64  `json:"categoryId"`
	Sort       int    `json:"sort"`
}

type httpDeleteSound struct {
	ID int64 `json:"id"`
}

type httpUpdateSound struct {
	httpDeleteSound
	httpCreateSound
}

func httpApiModSound(rctx *RequestContext) (any, error, error) {
	switch rctx.Req.Method {
	case http.MethodPost:
		{
			params := httpCreateSound{}
			if err := httpApiJsonRead(rctx.Req, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.Name == "" || params.Hash == "" || params.CategoryID == 0 {
				return nil, errMissingParam, errMissingParam
			}

			category, err := rctx.Services.Database.FetchCategoryByID(params.CategoryID)
			if err != nil {
				return nil, err, errDatabaseRead

			}

			if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			sound := database.Sound{Name: params.Name, Hash: params.Hash, CategoryID: params.CategoryID, Sort: params.Sort}
			if err := rctx.Services.Database.Save(&sound); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "sound": sound}, nil, nil
		}
	case http.MethodPut:
		{
			params := httpUpdateSound{}
			if err := httpApiJsonRead(rctx.Req, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.ID == 0 || params.Name == "" || params.Hash == "" || params.CategoryID == 0 {
				return nil, errMissingParam, errMissingParam
			}

			category, err := rctx.Services.Database.FetchCategoryByID(params.CategoryID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			sound, err := rctx.Services.Database.FetchSoundByID(params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			sound.Name = params.Name
			sound.Hash = params.Hash
			sound.CategoryID = params.CategoryID
			sound.Sort = params.Sort
			if err := rctx.Services.Database.Save(&sound); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "sound": sound}, nil, nil
		}
	case http.MethodDelete:
		{
			params := httpDeleteSound{}
			if err := httpApiJsonRead(rctx.Req, &params); err != nil {
				return nil, err, errInvalidParam
			}

			sound, err := rctx.Services.Database.FetchSoundByID(params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			category, err := rctx.Services.Database.FetchCategoryByID(sound.CategoryID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			if err := rctx.Services.Database.Delete(&sound); err != nil {
				return nil, err, errDatabaseWrite
			}

			return nil, nil, nil
		}
	}
	return nil, errInvalidMethod, errInvalidMethod
}

type httpSortSound struct {
	CategoryID int64   `json:"categoryId"`
	IDs        []int64 `json:"ids"`
}

func httpApiSortSound(rctx *RequestContext) (any, error, error) {
	params := httpSortSound{}
	if err := httpApiJsonRead(rctx.Req, &params); err != nil {
		return nil, err, errInvalidData
	}

	category, err := rctx.Services.Database.FetchCategoryByID(params.CategoryID)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	if err := checkUserGuild(rctx.Services.Database, rctx.User.ID, category.GuildID); err != nil {
		return nil, err, errNotAuthorized
	}

	if err := rctx.Services.Database.SortSounds(params.CategoryID, params.IDs...); err != nil {
		return nil, err, errDatabaseWrite
	}

	return nil, nil, nil
}

func httpApiUploadSound(rctx *RequestContext) (any, error, error) {
	// TODO more security to prevert user from uploading anything they want
	fileInfo, err := rctx.Services.Sound.Save(rctx.Req.Body)
	if err != nil {
		return nil, err, errSoundWrite
	}
	return map[string]any{"success": true, "hash": fileInfo.Hash}, nil, nil
}

type httpPlaySound struct {
	ID int64 `json:"id"`
}

func httpApiPlaySound(rctx *RequestContext) (any, error, error) {
	params := httpPlaySound{}
	if err := httpApiJsonRead(rctx.Req, &params); err != nil {
		return nil, err, errInvalidParam
	}

	sound, guildID, err := rctx.Services.Database.FetchSoundByIDAndUser(params.ID, rctx.User.ID)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	if sound.Hash != "" {
		soundReader, err := rctx.Services.Sound.Load(sound.Hash)
		if err != nil {
			return nil, err, errSoundRead
		}

		vs, err := rctx.Services.Discord.VoiceSession(guildID)
		if err != nil {
			return nil, err, errDiscordApi
		}

		channelID, err := rctx.Services.Discord.UserVoiceChannel(guildID, rctx.User.ID)
		if err != nil {
			return nil, err, errDiscordApi
		}
		if err := vs.Play(soundReader, channelID); err != nil {
			return nil, err, err
		}
	}

	return nil, nil, nil
}

type httpStopSounds struct {
	GuildID string `json:"guildId"`
}

func httpApiStopSounds(rctx *RequestContext) (any, error, error) {
	params := httpStopSounds{}
	if err := httpApiJsonRead(rctx.Req, &params); err != nil {
		return nil, err, errInvalidParam
	}

	vs, err := rctx.Services.Discord.VoiceSession(params.GuildID)
	if err != nil {
		return nil, err, errDiscordApi
	}
	vs.Stop()

	return nil, nil, nil
}

func handleHttp(services *Services, callback func(rctx *RequestContext) (string, error, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestIDCounter++

		logger := services.Logger.With().Int("requestID", requestIDCounter).Logger()
		rctx := RequestContext{
			ID:       requestIDCounter,
			Req:      r,
			RW:       w,
			Services: services,
			Logger:   &logger,
		}

		redirectURL, err, userErr := callback(&rctx)
		if err != nil {
			httpError(&rctx, err, userErr)
			return
		}

		if redirectURL != "" {
			logger.Info().Msgf("%d | %d | %s %s", rctx.ID, http.StatusTemporaryRedirect, r.Method, r.URL.Path)
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		}
	}
}

func httpError(rctx *RequestContext, err error, userErr error) {
	if userErr == nil {
		userErr = errUnknown
	}
	statusCode := errHttpStatusCodeMap[userErr]
	if statusCode == 0 {
		statusCode = 500
	}
	rctx.Logger.Warn().Err(err).Msgf("%d | %d | %s %s -- %s", rctx.ID, statusCode, rctx.Req.Method, rctx.Req.URL.Path, err.Error())
	rctx.RW.Header().Add("Content-Type", "text/plain")
	rctx.RW.WriteHeader(statusCode)
	rctx.RW.Write([]byte("ERROR: " + userErr.Error()))
}

func Serve(
	port int,
	databaseClient *database.Client,
	discordClient *discord.Client,
	discordAuthClient *disgoauth.Client,
	soundClient *sound.Client,
	logger *zerolog.Logger,
) error {
	services := Services{
		Discord:  discordClient,
		Auth:     discordAuthClient,
		Database: databaseClient,
		Sound:    soundClient,
		Logger:   logger,
	}

	// static
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("web"))))

	// auth
	http.HandleFunc("/login", handleHttp(&services, httpLogin))
	http.HandleFunc("/redirect", handleHttp(&services, httpLoginRedirect))

	// api
	http.HandleFunc("/api/me", handleHttpApi(&services, httpApiMe))
	http.HandleFunc("/api/list_user_guilds", handleHttpApi(&services, httpApiListGuilds))
	http.HandleFunc("/api/list_guild_categories", handleHttpApi(&services, httpApiListGuildCategories))
	http.HandleFunc("/api/list_guild_sounds", handleHttpApi(&services, httpApiListGuildSounds))
	http.HandleFunc("/api/list_guild_categories_and_sounds", handleHttpApi(&services, httpApiListGuildCategoriesAndSounds))
	http.HandleFunc("/api/category", handleHttpApi(&services, httpApiModCategory))
	http.HandleFunc("/api/sort_guild_categories", handleHttpApi(&services, httpApiSortCategory))
	http.HandleFunc("/api/sound", handleHttpApi(&services, httpApiModSound))
	http.HandleFunc("/api/sort_category_sounds", handleHttpApi(&services, httpApiSortSound))
	http.HandleFunc("/api/upload_sound", handleHttpApi(&services, httpApiUploadSound))
	http.HandleFunc("/api/play_sound", handleHttpApi(&services, httpApiPlaySound))
	http.HandleFunc("/api/stop_sounds", handleHttpApi(&services, httpApiStopSounds))

	logger.Info().Msgf("Start web server at http://0.0.0.0:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
