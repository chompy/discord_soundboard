package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type discordUserGuild struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner bool   `json:"owner"`
}

type UserPermissions uint

const (
	PermNone  UserPermissions = 0
	PermOwner UserPermissions = 1 << iota
	PermPlay
	PermAdminSound
	PermAdminUser
)

type UserGuild struct {
	ID          int64           `json:"id"`
	UserID      string          `json:"userId"`
	GuildID     string          `json:"guildId"`
	Permissions UserPermissions `json:"permissions"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
}

func (t UserGuild) name() string {
	return "user_guilds"
}

func (t UserGuild) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		guild_id TEXT NOT NULL,
		permissions INTEGER,
		created DATETIME,
		updated DATETIME,
		UNIQUE(user_id, guild_id)
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *UserGuild) id() any {
	return t.ID
}

func (t *UserGuild) save(conn *sql.DB) error {
	t.Updated = time.Now()
	if t.ID > 0 {
		stmt := fmt.Sprintf(`
		UPDATE %s
		SET permissions=?, updated=?
		WHERE id=?
		`, t.name())
		_, err := conn.Exec(stmt, t.Permissions, t.Updated, t.ID)
		return err
	}

	t.Created = t.Updated
	stmt := fmt.Sprintf(`
	INSERT OR IGNORE INTO %s
		(user_id, guild_id, permissions, created, updated) 
		VALUES(?, ?, ?, ?, ?)
	`, t.name())
	r, err := conn.Exec(stmt, t.UserID, t.GuildID, t.Permissions, t.Created, t.Updated)
	if err != nil {
		return err
	}
	t.ID, err = r.LastInsertId()
	return err
}

func (t *UserGuild) delete(conn *sql.DB) error {
	return deleteById(conn, t)
}

func userGuildsReadRows(rows *sql.Rows) ([]UserGuild, error) {
	out := make([]UserGuild, 0)
	for rows.Next() {
		r := UserGuild{}
		if err := rows.Scan(&r.ID, &r.UserID, &r.GuildID, &r.Permissions, &r.Created, &r.Updated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchUserGuildByID(ID int64) (UserGuild, error) {
	if c.conn == nil {
		return UserGuild{}, errDatabaseClosed
	}
	rows, err := fetchById(c.conn, &UserGuild{}, ID)
	if err != nil {
		return UserGuild{}, err
	}
	defer rows.Close()
	data, err := userGuildsReadRows(rows)
	if err != nil {
		return UserGuild{}, err
	}
	if len(data) == 0 {
		return UserGuild{}, errRecordNotFound
	}
	return data[0], nil
}

func (c *Client) FetchUserGuildsByUserID(userId string) ([]UserGuild, error) {
	if c.conn == nil {
		return nil, errDatabaseClosed
	}
	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE user_id = ?`, UserGuild{}.name())
	rows, err := c.conn.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return userGuildsReadRows(rows)
}

func (c *Client) UserHasGuild(userId string, guildId string) (bool, error) {
	if c.conn == nil {
		return false, errDatabaseClosed
	}
	stmt := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ? AND guild_id = ?`, UserGuild{}.name())
	rows, err := c.conn.Query(stmt, userId, guildId)
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

func (c *Client) ImportUserGuilds(reader io.Reader, userID string) error {
	if c.conn == nil {
		return errDatabaseClosed
	}

	c.logger.Info().Str("userID", userID).Msgf("Import guilds for user %s", userID)

	rawData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	discordGuilds := make([]discordUserGuild, 0)
	if err := json.Unmarshal(rawData, &discordGuilds); err != nil {
		return err
	}

	userGuilds, err := c.FetchUserGuildsByUserID(userID)
	if err != nil {
		return err
	}

	// remove user from guilds they are no longer in
	for _, userGuild := range userGuilds {
		hasGuild := false
		for _, discordGuild := range discordGuilds {
			hasGuild = discordGuild.ID == userGuild.GuildID
			if hasGuild {
				break
			}
		}
		if !hasGuild {
			c.logger.Info().Str("userID", userID).Str("guildID", userGuild.GuildID).Msgf("Remove user %s from guild %s", userID, userGuild.GuildID)
			if err := c.Delete(&userGuild); err != nil {
				return err
			}
		}
	}

	for _, discordGuild := range discordGuilds {
		c.logger.Info().Str("userID", userID).Str("guildID", discordGuild.ID).Msgf("Add user %s to guild %s (%s)", userID, discordGuild.ID, discordGuild.Name)

		permissions := PermPlay
		for _, userGuild := range userGuilds {
			if userGuild.GuildID == discordGuild.ID {
				permissions = userGuild.Permissions
			}
		}
		if discordGuild.Owner {
			permissions = PermOwner
		}

		gdb := Guild{ID: discordGuild.ID, Name: discordGuild.Name}
		if err := c.Save(&gdb); err != nil {
			return err
		}

		ug := UserGuild{GuildID: discordGuild.ID, UserID: userID, Permissions: permissions}
		if err := c.Save(&ug); err != nil {
			return err
		}
	}

	return nil
}
