package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Category struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	GuildID string    `json:"guildId"`
	Sort    int       `json:"sort"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func (t Category) name() string {
	return "categories"
}

func (t Category) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		guild_id TEXT,
		sort INTEGER,
		created DATETIME,
		updated DATETIME
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *Category) id() any {
	return t.ID
}

func (t *Category) save(conn *sql.DB) error {
	t.Updated = time.Now()
	if t.ID > 0 {
		stmt := fmt.Sprintf(`
		UPDATE %s
		SET name=?, guild_id=?, sort=?, updated=?
		WHERE id=?
		`, t.name())
		_, err := conn.Exec(stmt, t.Name, t.GuildID, t.Sort, t.Updated, t.ID)
		return err
	}

	t.Created = time.Now()
	stmt := fmt.Sprintf(`
	INSERT INTO %s (name, guild_id, sort, created, updated) 
	VALUES(?, ?, ?, ?, ?)
	`, t.name())
	r, err := conn.Exec(stmt, t.Name, t.GuildID, t.Sort, t.Created, t.Updated)
	if err != nil {
		return err
	}
	t.ID, err = r.LastInsertId()
	return err
}

func (t *Category) delete(conn *sql.DB) error {
	// delete sounds in category
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE category_id = ?`, Sound{}.name())
	_, err := conn.Exec(stmt, t.ID)
	if err != nil {
		return err
	}
	return deleteById(conn, t)
}

func catergoryReadRows(rows *sql.Rows) ([]Category, error) {
	out := make([]Category, 0)
	for rows.Next() {
		r := Category{}
		if err := rows.Scan(&r.ID, &r.Name, &r.GuildID, &r.Sort, &r.Created, &r.Updated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchCategoryByID(ID int64) (Category, error) {
	if c.conn == nil {
		return Category{}, errDatabaseClosed
	}
	rows, err := fetchById(c.conn, &Category{}, ID)
	if err != nil {
		return Category{}, err
	}
	data, err := catergoryReadRows(rows)
	if err != nil {
		return Category{}, err
	}
	if len(data) == 0 {
		return Category{}, errRecordNotFound
	}
	return data[0], nil
}

func (c *Client) FetchCategoriesByGuildID(guildID string) ([]Category, error) {
	if c.conn == nil {
		return nil, errDatabaseClosed
	}
	stmt := fmt.Sprintf("SELECT * FROM %s WHERE guild_id = ? ORDER BY sort, id ASC", Category{}.name())
	rows, err := c.conn.Query(stmt, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return catergoryReadRows(rows)
}

func (c *Client) SortCategories(guildID string, IDs ...int64) error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	tx, err := c.conn.Begin()
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(`UPDATE %s SET sort = ? WHERE guild_id = ?`, Category{}.name())
	if _, err := tx.Exec(stmt, 9999, guildID); err != nil {
		return err
	}
	for index, ID := range IDs {
		stmt = fmt.Sprintf(`UPDATE %s SET sort = ? WHERE id = ? AND guild_id = ?`, Category{}.name())
		if _, err := tx.Exec(stmt, index, ID, guildID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
