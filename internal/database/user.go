package database

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SessionToken string    `json:"-"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
}

func (t User) name() string {
	return "users"
}

func (t User) init(conn *sql.DB) error {
	stmt := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT,
		session_token TEXT,
		created DATETIME,
		updated DATETIME
	)
	`, t.name())
	_, err := conn.Exec(stmt)
	return err
}

func (t *User) id() any {
	return t.ID
}

func (t *User) save(conn *sql.DB) error {
	if t.Created.IsZero() {
		t.Created = time.Now()
	}
	t.Updated = time.Now()
	stmt := fmt.Sprintf(`
	INSERT INTO %s
		(id, name, session_token, created, updated) 
		VALUES(?, ?, ?, ?, ?) ON CONFLICT(id) 
		DO UPDATE SET name=excluded.name, session_token=excluded.session_token, updated=excluded.updated
	`, t.name())
	_, err := conn.Exec(stmt, t.ID, t.Name, t.SessionToken, t.Created, t.Updated)
	return err
}

func (t *User) delete(conn *sql.DB) error {
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE user_id = ?`, UserGuild{}.name())
	_, err := conn.Exec(stmt, t.ID)
	if err != nil {
		return err
	}
	return deleteById(conn, t)
}

func userReadRows(rows *sql.Rows) ([]User, error) {
	out := make([]User, 0)
	for rows.Next() {
		r := User{}
		if err := rows.Scan(&r.ID, &r.Name, &r.SessionToken, &r.Created, &r.Updated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchUserByID(ID string) (User, error) {
	if c.conn == nil {
		return User{}, errDatabaseClosed
	}
	rows, err := fetchById(c.conn, &User{}, ID)
	if err != nil {
		return User{}, err
	}
	data, err := userReadRows(rows)
	if err != nil {
		return User{}, err
	}
	if len(data) == 0 {
		return User{}, errRecordNotFound
	}
	return data[0], nil
}
