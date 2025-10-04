package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Guild struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func (t Guild) name() string {
	return "guilds"
}

func (t Guild) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		name TEXT,
		created DATETIME,
		updated DATETIME
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *Guild) id() any {
	return t.ID
}

func (t *Guild) save(conn *sql.DB) error {
	if t.Created.IsZero() {
		t.Created = time.Now()
	}
	t.Updated = time.Now()
	stmt := fmt.Sprintf(`
	INSERT INTO %s
		(id, name, created, updated) 
		VALUES(?, ?, ?, ?) ON CONFLICT(id) 
		DO UPDATE SET name=excluded.name, updated=excluded.updated
	`, t.name())
	_, err := conn.Exec(stmt, t.ID, t.Name, t.Created, t.Updated)
	return err
}

func (t *Guild) delete(conn *sql.DB) error {
	return deleteById(conn, t)
}

func guildReadRows(rows *sql.Rows) ([]Guild, error) {
	out := make([]Guild, 0)
	for rows.Next() {
		r := Guild{}
		if err := rows.Scan(&r.ID, &r.Name, &r.Created, &r.Updated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchGuildByID(ID string) (Guild, error) {
	if c.conn == nil {
		return Guild{}, errDatabaseClosed
	}
	out := Guild{}
	rows, err := fetchById(c.conn, &out, ID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	data, err := guildReadRows(rows)
	if err != nil {
		return Guild{}, err
	}
	if len(data) == 0 {
		return Guild{}, errRecordNotFound
	}
	return data[0], nil
}
