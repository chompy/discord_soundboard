package database

import (
	"database/sql"
	"fmt"
	"time"
)

type UserFavorite struct {
	ID      int64     `json:"id"`
	UserID  string    `json:"userId"`
	SoundID int64     `json:"soundId"`
	Created time.Time `json:"created"`
}

func (t UserFavorite) name() string {
	return "user_favorites"
}

func (t UserFavorite) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		sound_id INTEGER NOT NULL,
		created DATETIME,
		UNIQUE(user_id, sound_id)
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *UserFavorite) id() any {
	return t.ID
}

func (t *UserFavorite) save(conn *sql.DB) error {
	if t.ID > 0 {
		stmt := fmt.Sprintf(`
		UPDATE %s
		SET user_id=?, sound_id=?
		WHERE id=?
		`, t.name())
		_, err := conn.Exec(stmt, t.UserID, t.SoundID, t.ID)
		return err
	}

	t.Created = time.Now()
	stmt := fmt.Sprintf(`
	INSERT OR IGNORE INTO %s
		(user_id, sound_id, created) 
		VALUES(?, ?, ?)
	`, t.name())
	r, err := conn.Exec(stmt, t.UserID, t.SoundID, t.Created)
	if err != nil {
		return err
	}
	t.ID, err = r.LastInsertId()
	return err
}

func (t *UserFavorite) delete(conn *sql.DB) error {
	return deleteById(conn, t)
}

func userFavoritesReadRows(rows *sql.Rows) ([]UserFavorite, error) {
	out := make([]UserFavorite, 0)
	for rows.Next() {
		r := UserFavorite{}
		if err := rows.Scan(&r.ID, &r.UserID, &r.SoundID, &r.Created); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchUserFavoritesByUserID(userId string) ([]UserFavorite, error) {
	if c.conn == nil {
		return nil, errDatabaseClosed
	}
	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE user_id = ?`, UserFavorite{}.name())
	rows, err := c.conn.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return userFavoritesReadRows(rows)
}

func (c *Client) DeleteUserFavoriteByUserIDAndSoundID(userId string, soundId int64) error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE user_id = ? AND sound_id = ?`, UserFavorite{}.name())
	_, err := c.conn.Exec(stmt, userId, soundId)
	return err
}
