package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type ApiRequest struct {
	ID   int
	App  *App
	R    *http.Request
	W    http.ResponseWriter
	User User
	DB   *sql.DB
}

var requestIdCounter = 0

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

func httpApiError(r *ApiRequest, err error, userErr error) {
	if userErr == nil {
		userErr = errUnknown
	}
	statusCode := errHttpStatusCodeMap[userErr]
	if statusCode == 0 {
		statusCode = 500
	}
	log.Printf("> R:%05d | %d - %s %s | ERROR: %s", r.ID, statusCode, r.R.Method, r.R.URL.Path, err)
	httpApiJsonWrite(r.W, map[string]any{"success": false, "error": userErr.Error(), "isLoggedIn": r.User.ID != ""}, statusCode)
}

func handleHttpApi(app *App, callback func(r *ApiRequest) (any, error, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestIdCounter++
		apiReq := ApiRequest{ID: requestIdCounter, App: app, W: w, R: r}

		//log.Println("> R:%05d | %s %s", out.ID, r.Method, r.URL.Path)

		var err error
		apiReq.DB, err = databaseOpen()
		if err != nil {
			httpApiError(&apiReq, err, errDatabaseOpen)
			return
		}
		defer apiReq.DB.Close()

		apiReq.User, err = checkUser(apiReq.DB, r)
		if err != nil {
			httpApiError(&apiReq, err, errNotAuthenticated)
			return
		}

		resp, err, userErr := callback(&apiReq)
		if err != nil {
			httpApiError(&apiReq, err, userErr)
		}

		if resp == nil {
			resp = map[string]any{"success": true}
		}
		httpApiJsonWrite(w, resp, http.StatusOK)
		log.Printf("> R:%05d | 200 - %s %s | -", apiReq.ID, r.Method, r.URL.Path)
	}
}

func httpApiMe(r *ApiRequest) (any, error, error) {
	return map[string]any{"success": true, "user": r.User}, nil, nil
}

func httpApiListGuilds(r *ApiRequest) (any, error, error) {
	userGuilds, err := r.App.Discord.UserAvailableGuilds(r.DB, r.User.ID)
	if err != nil {
		return nil, err, errDiscordApi
	}
	return map[string]any{"success": true, "guilds": userGuilds}, nil, nil
}

func httpApiListGuildCategories(r *ApiRequest) (any, error, error) {
	guildId := r.R.URL.Query().Get("guild")
	if guildId == "" {
		return nil, errMissingParam, errMissingParam
	}
	if err := checkUserGuild(r.DB, r.User.ID, guildId); err != nil {
		return nil, err, errNotAuthorized
	}

	categories, err := databaseFetchCategoriesByGuildID(r.DB, guildId)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	return map[string]any{"success": true, "categories": categories}, nil, nil
}

func httpApiListGuildSounds(r *ApiRequest) (any, error, error) {
	guildId := r.R.URL.Query().Get("guild")
	if guildId == "" {
		return nil, errMissingParam, errMissingParam
	}

	if err := checkUserGuild(r.DB, r.User.ID, guildId); err != nil {
		return nil, err, errNotAuthorized
	}

	sounds, err := databaseFetchSoundsByGuildID(r.DB, guildId)
	if err != nil {
		return nil, err, errDatabaseRead
	}
	return map[string]any{"success": true, "sounds": sounds}, nil, nil
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

func httpApiModCategory(r *ApiRequest) (any, error, error) {
	switch r.R.Method {
	case http.MethodPost:
		{
			params := httpCreateCategory{}
			if err := httpApiJsonRead(r.R, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.Name == "" || params.GuildID == "" {
				return nil, errMissingParam, errMissingParam
			}

			if err := checkUserGuild(r.DB, r.User.ID, params.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			category := Category{Name: params.Name, GuildID: params.GuildID, Sort: params.Sort}
			if err := category.Save(r.DB); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "category": category}, nil, nil
		}
	case http.MethodPut:
		{
			params := httpUpdateCategory{}
			if err := httpApiJsonRead(r.R, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.ID == 0 || params.Name == "" || params.GuildID == "" {
				return nil, errMissingParam, errMissingParam
			}

			if err := checkUserGuild(r.DB, r.User.ID, params.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			category, err := databaseFetchCategoryByID(r.DB, params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			category.Name = params.Name
			category.GuildID = params.GuildID
			category.Sort = params.Sort
			if err := category.Save(r.DB); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "category": category}, nil, nil
		}
	case http.MethodDelete:
		{
			params := httpDeleteCategory{}
			if err := httpApiJsonRead(r.R, &params); err != nil {
				return nil, err, errInvalidParam
			}

			category, err := databaseFetchCategoryByID(r.DB, params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			if err := checkUserGuild(r.DB, r.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			if err := category.Delete(r.DB); err != nil {
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

func httpApiSortCategory(r *ApiRequest) (any, error, error) {
	params := httpSortCategory{}
	if err := httpApiJsonRead(r.R, &params); err != nil {
		return nil, err, errInvalidParam
	}

	if err := databaseSortCategories(r.DB, params.GuildID, params.IDs...); err != nil {
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

func httpApiModSound(r *ApiRequest) (any, error, error) {
	switch r.R.Method {
	case http.MethodPost:
		{
			params := httpCreateSound{}
			if err := httpApiJsonRead(r.R, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.Name == "" || params.Hash == "" || params.CategoryID == 0 {
				return nil, errMissingParam, errMissingParam
			}

			category, err := databaseFetchCategoryByID(r.DB, params.CategoryID)
			if err != nil {
				return nil, err, errDatabaseRead

			}

			if err := checkUserGuild(r.DB, r.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			sound := Sound{Name: params.Name, Hash: params.Hash, CategoryID: params.CategoryID, Sort: params.Sort}
			if err := sound.Save(r.DB); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "sound": sound}, nil, nil
		}
	case http.MethodPut:
		{
			params := httpUpdateSound{}
			if err := httpApiJsonRead(r.R, &params); err != nil {
				return nil, err, errInvalidParam
			}

			if params.ID == 0 || params.Name == "" || params.Hash == "" || params.CategoryID == 0 {
				return nil, errMissingParam, errMissingParam
			}

			category, err := databaseFetchCategoryByID(r.DB, params.CategoryID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			sound, err := databaseFetchSoundByID(r.DB, params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			if err := checkUserGuild(r.DB, r.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			sound.Name = params.Name
			sound.Hash = params.Hash
			sound.CategoryID = params.CategoryID
			sound.Sort = params.Sort
			if err := sound.Save(r.DB); err != nil {
				return nil, err, errDatabaseWrite
			}

			return map[string]any{"success": true, "sound": sound}, nil, nil
		}
	case http.MethodDelete:
		{
			params := httpDeleteSound{}
			if err := httpApiJsonRead(r.R, &params); err != nil {
				return nil, err, errInvalidParam
			}

			sound, err := databaseFetchSoundByID(r.DB, params.ID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			category, err := databaseFetchCategoryByID(r.DB, sound.CategoryID)
			if err != nil {
				return nil, err, errDatabaseRead
			}

			if err := checkUserGuild(r.DB, r.User.ID, category.GuildID); err != nil {
				return nil, err, errNotAuthorized
			}

			if err := sound.Delete(r.DB); err != nil {
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

func httpApiSortSound(r *ApiRequest) (any, error, error) {
	params := httpSortSound{}
	if err := httpApiJsonRead(r.R, &params); err != nil {
		return nil, err, errInvalidData
	}

	category, err := databaseFetchCategoryByID(r.DB, params.CategoryID)
	if err != nil {
		return nil, err, errDatabaseRead
	}

	if err := checkUserGuild(r.DB, r.User.ID, category.GuildID); err != nil {
		return nil, err, errNotAuthorized
	}

	if err := databaseSortSounds(r.DB, params.CategoryID, params.IDs...); err != nil {
		return nil, err, errDatabaseWrite
	}

	return nil, nil, nil
}

func httpApiUploadSound(r *ApiRequest) (any, error, error) {
	// TODO more security to prevert user from uploading anything they want

	soundReader := NewSoundReader(r.R.Body)
	hash, err := soundReader.Save()
	if err != nil {
		return nil, err, errSoundWrite
	}

	return map[string]any{"success": true, "hash": hash}, nil, nil
}

type httpPlaySound struct {
	ID int64 `json:"id"`
}

func httpApiPlaySound(r *ApiRequest) (any, error, error) {
	params := httpPlaySound{}
	if err := httpApiJsonRead(r.R, &params); err != nil {
		return nil, err, errInvalidParam
	}

	sound, guildId, err := databaseFetchSoundByIDAndUser(r.DB, params.ID, r.User.ID)
	if err != nil {
		return nil, err, errDatabaseRead
	}
	if sound.Hash != "" {
		soundReader, err := NewSoundReaderFromStorage(sound.Hash)
		if err != nil {
			return nil, err, errSoundRead
		}

		vs := r.App.Discord.VoiceSession(guildId)
		channelId, err := r.App.Discord.UserVoiceChannel(guildId, r.User.ID)
		if err != nil {
			return nil, err, errDiscordApi
		}
		if err := vs.Play(soundReader, channelId); err != nil {
			return nil, err, err
		}
	}

	return nil, nil, nil
}

type httpStopSounds struct {
	GuildID string `json:"guildId"`
}

func httpApiStopSounds(r *ApiRequest) (any, error, error) {
	params := httpStopSounds{}
	if err := httpApiJsonRead(r.R, &params); err != nil {
		return nil, err, errInvalidParam
	}

	vs := r.App.Discord.VoiceSession(params.GuildID)
	vs.Stop()

	return nil, nil, nil
}

func httpError(w http.ResponseWriter, err error) {
	statusCode := errHttpStatusCodeMap[err]
	if statusCode == 0 {
		statusCode = 500
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte("ERROR: " + err.Error()))
}

func RunWebServer(app *App) error {

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("web"))))

	http.HandleFunc("/login", httpLogin)
	http.HandleFunc("/redirect", httpLoginRedirect)

	http.HandleFunc("/api/me", handleHttpApi(app, httpApiMe))
	http.HandleFunc("/api/list_user_guilds", handleHttpApi(app, httpApiListGuilds))
	http.HandleFunc("/api/list_guild_categories", handleHttpApi(app, httpApiListGuildCategories))
	http.HandleFunc("/api/list_guild_sounds", handleHttpApi(app, httpApiListGuildSounds))
	http.HandleFunc("/api/category", handleHttpApi(app, httpApiModCategory))
	http.HandleFunc("/api/sort_guild_categories", handleHttpApi(app, httpApiSortCategory))
	http.HandleFunc("/api/sound", handleHttpApi(app, httpApiModSound))
	http.HandleFunc("/api/sort_category_sounds", handleHttpApi(app, httpApiSortSound))
	http.HandleFunc("/api/upload_sound", handleHttpApi(app, httpApiUploadSound))
	http.HandleFunc("/api/play_sound", handleHttpApi(app, httpApiPlaySound))
	http.HandleFunc("/api/stop_sounds", handleHttpApi(app, httpApiStopSounds))

	log.Println("> Start web server at http://localhost:8081")
	return http.ListenAndServe(":8081", nil)
}
