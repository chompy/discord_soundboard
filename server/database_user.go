package main

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SessionToken string    `json:"-"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
}

func databaseCreateUserTable(db *sql.DB) error {
	log.Println("  - Create users table")
	stmt := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT,
		session_token TEXT,
		created DATETIME,
		updated DATETIME
	)
	`
	_, err := db.Exec(stmt)
	return err
}

func databaseFetchUserByID(db *sql.DB, ID string) (User, error) {
	user := User{}
	stmt := "SELECT * FROM users WHERE id = ?"
	rows, err := db.Query(stmt, ID)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.SessionToken, &user.Created, &user.Updated)
		return user, err
	}
	return user, nil
}

func (u *User) Save(db *sql.DB) error {
	if u.Created.IsZero() {
		u.Created = time.Now()
	}
	u.Updated = time.Now()
	stmt := `
	INSERT INTO users
		(id, name, session_token, created, updated) 
		VALUES(?, ?, ?, ?, ?) ON CONFLICT(id) 
		DO UPDATE SET name=excluded.name, session_token=excluded.session_token, updated=excluded.updated
	`
	_, err := db.Exec(stmt, u.ID, u.Name, u.SessionToken, u.Created, u.Updated)
	return err
}
