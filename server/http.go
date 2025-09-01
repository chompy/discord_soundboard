package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func httpApiInit(w http.ResponseWriter, r *http.Request) (*sql.DB, User) {
	db, err := databaseOpen()
	if err != nil {
		httpApiError(w, err)
		return nil, User{}
	}
	user, err := checkUser(db, r)
	if err != nil {
		db.Close()
		httpApiError(w, err)
		return nil, User{}
	}
	return db, user
}

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

func httpApiError(w http.ResponseWriter, err error) {
	log.Println("> ERROR: ", err)
	httpApiJsonWrite(w, map[string]any{"success": false, "error": err.Error()}, http.StatusInternalServerError)
}

func httpApiSuccess(w http.ResponseWriter) {
	httpApiJsonWrite(w, map[string]any{"success": true}, http.StatusOK)
}

func httpApiMe(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db, user := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		httpApiJsonWrite(w, map[string]any{"success": true, "user": user}, http.StatusOK)
	}
}

func httpApiListGuilds(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db, user := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		userGuilds, err := app.Discord.UserAvailableGuilds(db, user.ID)
		if err != nil {
			httpApiError(w, err)
			return
		}

		httpApiJsonWrite(w, map[string]any{"success": true, "guilds": userGuilds}, http.StatusOK)
	}
}

func httpApiListGuildCategories(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		guildId := r.URL.Query().Get("guild")
		if guildId == "" {
			httpApiError(w, errMissingParam)
			return
		}

		db, user := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		if err := checkUserGuild(db, user.ID, guildId); err != nil {
			httpApiError(w, err)
			return
		}

		categories, err := databaseFetchCategoriesByGuildID(db, guildId)
		if err != nil {
			httpApiError(w, err)
			return
		}
		httpApiJsonWrite(w, map[string]any{"success": true, "categories": categories}, http.StatusOK)
	}
}

func httpApiListGuildSounds(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		guildId := r.URL.Query().Get("guild")
		if guildId == "" {
			httpApiError(w, errMissingParam)
			return
		}

		db, user := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		if err := checkUserGuild(db, user.ID, guildId); err != nil {
			httpApiError(w, err)
			return
		}

		sounds, err := databaseFetchSoundsByGuildID(db, guildId)
		if err != nil {
			httpApiError(w, err)
			return
		}
		httpApiJsonWrite(w, map[string]any{"success": true, "sounds": sounds}, http.StatusOK)
	}
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

func httpApiModCategory(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		db, user := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		switch r.Method {
		case http.MethodPost:
			{
				params := httpCreateCategory{}
				if err := httpApiJsonRead(r, &params); err != nil {
					httpApiError(w, err)
					return
				}

				if params.Name == "" || params.GuildID == "" {
					httpApiError(w, errMissingParam)
					return
				}

				if err := checkUserGuild(db, user.ID, params.GuildID); err != nil {
					httpApiError(w, err)
					return
				}

				category := Category{Name: params.Name, GuildID: params.GuildID, Sort: params.Sort}
				if err := category.Save(db); err != nil {
					httpApiError(w, err)
					return
				}

				httpApiJsonWrite(w, map[string]any{"success": true, "category": category}, http.StatusOK)
				return
			}
		case http.MethodPut:
			{
				params := httpUpdateCategory{}
				if err := httpApiJsonRead(r, &params); err != nil {
					httpApiError(w, err)
					return
				}

				if params.ID == 0 || params.Name == "" || params.GuildID == "" {
					httpApiError(w, errMissingParam)
					return
				}

				if err := checkUserGuild(db, user.ID, params.GuildID); err != nil {
					httpApiError(w, err)
					return
				}

				category, err := databaseFetchCategoryByID(db, params.ID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				category.Name = params.Name
				category.GuildID = params.GuildID
				category.Sort = params.Sort
				if err := category.Save(db); err != nil {
					httpApiError(w, err)
					return
				}

				httpApiJsonWrite(w, map[string]any{"success": true, "category": category}, http.StatusOK)
				return
			}
		case http.MethodDelete:
			{
				params := httpDeleteCategory{}
				if err := httpApiJsonRead(r, &params); err != nil {
					httpApiError(w, err)
					return
				}

				category, err := databaseFetchCategoryByID(db, params.ID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				if err := checkUserGuild(db, user.ID, category.GuildID); err != nil {
					httpApiError(w, err)
					return
				}

				if err := category.Delete(db); err != nil {
					httpApiError(w, err)
					return
				}

				httpApiSuccess(w)
				return
			}
		}
		httpApiError(w, errInvalidMethod)
	}
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

func httpApiModSound(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		db, user := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		switch r.Method {
		case http.MethodPost:
			{
				params := httpCreateSound{}
				if err := httpApiJsonRead(r, &params); err != nil {
					httpApiError(w, err)
					return
				}

				if params.Name == "" || params.Hash == "" || params.CategoryID == 0 {
					httpApiError(w, errMissingParam)
					return
				}

				category, err := databaseFetchCategoryByID(db, params.CategoryID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				if err := checkUserGuild(db, user.ID, category.GuildID); err != nil {
					httpApiError(w, err)
					return
				}

				sound := Sound{Name: params.Name, Hash: params.Hash, CategoryID: params.CategoryID, Sort: params.Sort}
				if err := sound.Save(db); err != nil {
					httpApiError(w, err)
					return
				}

				httpApiJsonWrite(w, map[string]any{"success": true, "sound": sound}, http.StatusOK)
				return
			}
		case http.MethodPut:
			{
				params := httpUpdateSound{}
				if err := httpApiJsonRead(r, &params); err != nil {
					httpApiError(w, err)
					return
				}

				if params.ID == 0 || params.Name == "" || params.Hash == "" || params.CategoryID == 0 {
					httpApiError(w, errMissingParam)
					return
				}

				category, err := databaseFetchCategoryByID(db, params.CategoryID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				sound, err := databaseFetchSoundByID(db, params.ID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				if err := checkUserGuild(db, user.ID, category.GuildID); err != nil {
					httpApiError(w, err)
					return
				}

				sound.Name = params.Name
				sound.Hash = params.Hash
				sound.CategoryID = params.CategoryID
				sound.Sort = params.Sort
				if err := sound.Save(db); err != nil {
					httpApiError(w, err)
					return
				}

				httpApiJsonWrite(w, map[string]any{"success": true, "sound": sound}, http.StatusOK)
				return
			}
		case http.MethodDelete:
			{
				params := httpDeleteSound{}
				if err := httpApiJsonRead(r, &params); err != nil {
					httpApiError(w, err)
					return
				}

				sound, err := databaseFetchSoundByID(db, params.ID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				category, err := databaseFetchCategoryByID(db, sound.CategoryID)
				if err != nil {
					httpApiError(w, err)
					return
				}

				if err := checkUserGuild(db, user.ID, category.GuildID); err != nil {
					httpApiError(w, err)
					return
				}

				if err := sound.Delete(db); err != nil {
					httpApiError(w, err)
					return
				}

				httpApiSuccess(w)
				return
			}
		}
		httpApiError(w, errInvalidMethod)
	}
}

func httpApiUploadSound(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db, _ := httpApiInit(w, r)
		if db == nil {
			return
		}
		defer db.Close()

		// TODO more security to prevert user from uploading anything they want

		soundReader := NewSoundReader(r.Body)
		hash, err := soundReader.Save()
		if err != nil {
			httpApiError(w, err)
			return
		}

		httpApiJsonWrite(w, map[string]any{"success": true, "hash": hash}, http.StatusOK)
	}
}

func RunWebServer(app *App) error {

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("web"))))

	http.HandleFunc("/login", httpLogin)
	http.HandleFunc("/redirect", httpLoginRedirect)

	http.HandleFunc("/api/me", httpApiMe(app))
	http.HandleFunc("/api/list_user_guilds", httpApiListGuilds(app))
	http.HandleFunc("/api/list_guild_categories", httpApiListGuildCategories(app))
	http.HandleFunc("/api/list_guild_sounds", httpApiListGuildSounds(app))
	http.HandleFunc("/api/category", httpApiModCategory(app))
	http.HandleFunc("/api/sound", httpApiModSound(app))
	http.HandleFunc("/api/upload_sound", httpApiUploadSound(app))

	log.Println("> Start web server at http://localhost:8081")
	return http.ListenAndServe(":8081", nil)
}
