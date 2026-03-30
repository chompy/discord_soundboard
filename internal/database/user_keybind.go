package database

import (
	"database/sql"
	"fmt"
	"time"
)

type UserKeybind struct {
	ID      int64     `json:"id"`
	UserID  string    `json:"userId"`
	SoundID int64     `json:"soundId"`
	Key     string    `json:"key"`
	Created time.Time `json:"created"`
}

func (t UserKeybind) name() string {
	return "user_keybinds"
}

func (t UserKeybind) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		sound_id INTEGER NOT NULL,
		key TEXT NOT NULL,
		created DATETIME,
		UNIQUE(user_id, key)
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *UserKeybind) id() any {
	return t.ID
}

func (t *UserKeybind) save(conn *sql.DB) error {
	if t.ID > 0 {
		stmt := fmt.Sprintf(`
		UPDATE %s
		SET user_id=?, sound_id=?, key=?
		WHERE id=?
		`, t.name())
		_, err := conn.Exec(stmt, t.UserID, t.SoundID, t.Key, t.ID)
		return err
	}

	t.Created = time.Now()
	stmt := fmt.Sprintf(`
	INSERT OR IGNORE INTO %s
		(user_id, sound_id, key, created) 
		VALUES(?, ?, ?, ?)
	`, t.name())
	r, err := conn.Exec(stmt, t.UserID, t.SoundID, t.Key, t.Created)
	if err != nil {
		return err
	}
	t.ID, err = r.LastInsertId()
	return err
}

func (t *UserKeybind) delete(conn *sql.DB) error {
	return deleteById(conn, t)
}

func userKeybindReadRows(rows *sql.Rows) ([]UserKeybind, error) {
	out := make([]UserKeybind, 0)
	for rows.Next() {
		r := UserKeybind{}
		if err := rows.Scan(&r.ID, &r.UserID, &r.SoundID, &r.Key, &r.Created); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchUserKeybindsByUserID(userId string) ([]UserKeybind, error) {
	if c.conn == nil {
		return nil, errDatabaseClosed
	}
	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE user_id = ? GROUP BY key ORDER BY id ASC`, UserKeybind{}.name())
	rows, err := c.conn.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return userKeybindReadRows(rows)
}

func (c *Client) DeleteUserKeybindByUserIDAndSoundID(userId string, soundId int64) error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE user_id = ? AND sound_id = ?`, UserKeybind{}.name())
	_, err := c.conn.Exec(stmt, userId, soundId)
	return err
}

func (c *Client) DeleteUserKeybindByUserIDAndKey(userId string, key string) error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE user_id = ? AND key = ?`, UserKeybind{}.name())
	_, err := c.conn.Exec(stmt, userId, key)
	return err
}
