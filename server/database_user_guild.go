package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"time"
)

type UserPermissions uint

const (
	PermPlaySound UserPermissions = 1 << iota
	PermAddSound
	PermDelSound
)

type UserGuild struct {
	ID          int64           `json:"id"`
	UserID      string          `json:"userId"`
	GuildID     string          `json:"guildId"`
	Permissions UserPermissions `json:"permissions"`
	Guild       Guild           `json:"guild"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
}

func databaseCreateUserGuildTable(db *sql.DB) error {
	log.Println("  - Create user_guilds table")
	stmt := `
	CREATE TABLE IF NOT EXISTS user_guilds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		guild_id TEXT NOT NULL,
		permissions INTEGER,
		created DATETIME,
		updated DATETIME,
		UNIQUE(user_id, guild_id)
	)
	`
	_, err := db.Exec(stmt)
	return err
}

func DatabaseGetUserGuildsByUserID(db *sql.DB, userId string) ([]UserGuild, error) {
	out := make([]UserGuild, 0)

	stmt := `
	SELECT ug.id, ug.user_id, ug.guild_id, ug.permissions, ug.created, ug.updated, g.id, g.name, g.created, g.updated
	FROM user_guilds AS ug
	INNER JOIN guilds as g
	ON ug.guild_id = g.id
	WHERE ug.user_id = ?
	`
	rows, err := db.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userGuild := UserGuild{}
		if err := rows.Scan(
			&userGuild.ID, &userGuild.UserID, &userGuild.GuildID, &userGuild.Permissions, &userGuild.Created, &userGuild.Updated,
			&userGuild.Guild.ID, &userGuild.Guild.Name, &userGuild.Guild.Created, &userGuild.Guild.Updated); err != nil {
			return nil, err
		}
		out = append(out, userGuild)
	}

	return out, nil
}

func DatabaseDeleteUserGuildByID(db *sql.DB, ID int64) error {
	stmt := `DELETE FROM user_guilds WHERE id = ?`
	_, err := db.Exec(stmt, ID)
	return err
}

func DatabaseUserHasGuild(db *sql.DB, userId string, guildId string) (bool, error) {
	stmt := `
	SELECT COUNT(*) FROM user_guilds
	WHERE user_id = ? AND guild_id = ?
	`
	rows, err := db.Query(stmt, userId, guildId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		count := 0
		return count > 0, rows.Scan(&count)
	}
	return false, nil
}

func (u *UserGuild) Save(db *sql.DB) error {
	u.Updated = time.Now()
	if u.ID > 0 {
		stmt := `
		UPDATE user_guilds
		SET permissions=?, updated=?
		WHERE id=?
		`
		_, err := db.Exec(stmt, u.Permissions, u.Updated, u.ID)
		return err
	}

	u.Created = u.Updated
	stmt := `
	INSERT OR IGNORE INTO user_guilds 
		(user_id, guild_id, permissions, created, updated) 
		VALUES(?, ?, ?, ?, ?)
	`
	r, err := db.Exec(stmt, u.UserID, u.GuildID, u.Permissions, u.Created, u.Updated)
	if err != nil {
		return err
	}
	u.ID, err = r.LastInsertId()
	return err
}

func (u *UserGuild) Delete(db *sql.DB) error {
	return DatabaseDeleteUserGuildByID(db, u.ID)
}

func databaseImportUserGuilds(db *sql.DB, reader io.Reader, userId string) error {
	// TODO: add clean up process to remove user from guilds they are no longer in
	rawData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	data := make([]map[string]interface{}, 0)

	if err := json.Unmarshal(rawData, &data); err != nil {
		return err
	}

	for _, guild := range data {
		guildId, guildOk := guild["id"].(string)
		if !guildOk || guildId == "" {
			return errInvalidData
		}
		guildName, nameOk := guild["name"].(string)
		if !nameOk || guildName == "" {
			return errInvalidData
		}

		log.Printf("> Add user #%s to guild #%s (%s).", userId, guildId, guildName)

		gdb := Guild{ID: guildId, Name: guildName}
		if err := gdb.Save(db); err != nil {
			return err
		}

		ug := UserGuild{GuildID: guildId, UserID: userId}
		if err := ug.Save(db); err != nil {
			return err
		}
	}

	return nil

}
