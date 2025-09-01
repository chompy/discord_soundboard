package main

import (
	"database/sql"
	"log"
	"time"
)

type Guild struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func databaseCreateGuildTable(db *sql.DB) error {
	log.Println("  - Create guilds table")
	stmt := `
	CREATE TABLE IF NOT EXISTS guilds (
		id TEXT PRIMARY KEY,
		name TEXT,
		created DATETIME,
		updated DATETIME
	)
	`
	_, err := db.Exec(stmt)
	return err
}

func (g *Guild) Save(db *sql.DB) error {
	if g.Created.IsZero() {
		g.Created = time.Now()
	}
	g.Updated = time.Now()
	stmt := `
	INSERT INTO guilds
		(id, name, created, updated) 
		VALUES(?, ?, ?, ?) ON CONFLICT(id) 
		DO UPDATE SET name=excluded.name, updated=excluded.updated
	`
	_, err := db.Exec(stmt, g.ID, g.Name, g.Created, g.Updated)
	return err
}
