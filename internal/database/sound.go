package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Sound struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Hash       string    `json:"hash"`
	CategoryID int64     `json:"categoryId"`
	Sort       int       `json:"sort"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}

func (t Sound) name() string {
	return "sounds"
}

func (t Sound) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		hash TEXT,
		category_id INTEGER,
		sort INTEGER,
		created DATETIME,
		updated DATETIME
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *Sound) id() any {
	return t.ID
}

func (t *Sound) save(conn *sql.DB) error {
	t.Updated = time.Now()
	if t.ID > 0 {
		stmt := fmt.Sprintf(`
		UPDATE %s
		SET name=?, hash=?, category_id=?, sort=?, updated=?
		WHERE id=?
		`, t.name())
		_, err := conn.Exec(stmt, t.Name, t.Hash, t.CategoryID, t.Sort, t.Updated, t.ID)
		return err
	}

	t.Created = time.Now()
	stmt := fmt.Sprintf(`
	INSERT INTO %s (name, hash, category_id, sort, created, updated) 
	VALUES(?, ?, ?, ?, ?, ?)
	`, t.name())
	r, err := conn.Exec(stmt, t.Name, t.Hash, t.CategoryID, t.Sort, t.Created, t.Updated)
	if err != nil {
		return err
	}
	t.ID, err = r.LastInsertId()
	return err
}

func (t *Sound) delete(conn *sql.DB) error {
	return deleteById(conn, t)
}

func soundReadRows(rows *sql.Rows) ([]Sound, error) {
	out := make([]Sound, 0)
	for rows.Next() {
		r := Sound{}
		if err := rows.Scan(&r.ID, &r.Name, &r.Hash, &r.CategoryID, &r.Sort, &r.Created, &r.Updated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchSoundByID(ID int64) (Sound, error) {
	if c.conn == nil {
		return Sound{}, errDatabaseClosed
	}
	rows, err := fetchById(c.conn, &Sound{}, ID)
	if err != nil {
		return Sound{}, err
	}
	data, err := soundReadRows(rows)
	if err != nil {
		return Sound{}, err
	}
	if len(data) == 0 {
		return Sound{}, errRecordNotFound
	}
	return data[0], nil
}

func (c *Client) FetchSoundsByHash(hash string) ([]Sound, error) {
	if c.conn == nil {
		return nil, errDatabaseClosed
	}
	stmt := fmt.Sprintf(`SELECT * FROM %s WHERE hash = ?`, Sound{}.name())
	rows, err := c.conn.Query(stmt, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return soundReadRows(rows)
}

func (c *Client) FetchSoundsByGuildID(guildID string) ([]Sound, error) {
	if c.conn == nil {
		return nil, errDatabaseClosed
	}
	stmt := fmt.Sprintf(`
	SELECT s.id, s.name, s.hash, s.category_id, s.sort, s.created, s.updated FROM %s AS s
	INNER JOIN %s AS c ON s.category_id = c.id
	WHERE c.guild_id = ? ORDER BY s.sort, s.id ASC
	`, Sound{}.name(), Category{}.name())
	rows, err := c.conn.Query(stmt, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return soundReadRows(rows)
}

func (c *Client) SortSounds(categoryID int64, IDs ...int64) error {
	if c.conn == nil {
		return errDatabaseClosed
	}

	tx, err := c.conn.Begin()
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(`UPDATE %s SET sort = ? WHERE category_id = ?`, Sound{}.name())
	if _, err := tx.Exec(stmt, 9999, categoryID); err != nil {
		return err
	}
	for index, ID := range IDs {
		stmt = fmt.Sprintf(`UPDATE %s SET sort = ? WHERE id = ? AND category_id = ?`, Sound{}.name())
		if _, err := tx.Exec(stmt, index, ID, categoryID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) FetchSoundByIDAndUser(ID int64, userId string) (Sound, string, error) {
	if c.conn == nil {
		return Sound{}, "", errDatabaseClosed
	}
	stmt := fmt.Sprintf(`
	SELECT s.id, s.name, s.hash, s.category_id, s.sort, s.created, s.updated, c.guild_id
	FROM %s AS s
	INNER JOIN %s AS c ON s.category_id = c.id
	INNER JOIN %s AS u ON c.guild_id = u.guild_id 
	WHERE u.user_id = ? AND s.id = ?
	`, Sound{}.name(), Category{}.name(), UserGuild{}.name())
	rows, err := c.conn.Query(stmt, userId, ID)
	if err != nil {
		return Sound{}, "", err
	}
	defer rows.Close()

	for rows.Next() {
		sound := Sound{}
		guildId := ""
		err = rows.Scan(&sound.ID, &sound.Name, &sound.Hash, &sound.CategoryID, &sound.Sort, &sound.Created, &sound.Updated, &guildId)
		return sound, guildId, err
	}
	return Sound{}, "", nil
}
