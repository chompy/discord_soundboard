package main

import (
	"database/sql"
	"log"
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

func databaseCreateSoundTable(db *sql.DB) error {
	log.Println("  - Create sounds table")
	stmt := `
	CREATE TABLE IF NOT EXISTS sounds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		hash TEXT,
		category_id INTEGER,
		sort INTEGER,
		created DATETIME,
		updated DATETIME
	)
	`
	_, err := db.Exec(stmt)
	return err
}

func databaseFetchSoundByID(db *sql.DB, ID int64) (Sound, error) {
	sound := Sound{}

	stmt := `SELECT * FROM sounds WHERE id = ? ORDER BY sort, id ASC`
	rows, err := db.Query(stmt, ID)
	if err != nil {
		return sound, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&sound.ID, &sound.Name, &sound.Hash, &sound.CategoryID, &sound.Sort, &sound.Created, &sound.Updated)
		return sound, err
	}
	return sound, nil
}

func databaseFetchSoundsByGuildID(db *sql.DB, guildID string) ([]Sound, error) {
	out := make([]Sound, 0)

	stmt := `
	SELECT s.id, s.name, s.hash, s.category_id, s.sort, s.created, s.updated
	FROM sounds AS s
	INNER JOIN categories as c
	ON s.category_id = c.id
	WHERE c.guild_id = ?
	ORDER BY s.sort, s.id, c.sort, c.id ASC
	`
	rows, err := db.Query(stmt, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		sound := Sound{}
		if err := rows.Scan(&sound.ID, &sound.Name, &sound.Hash, &sound.CategoryID, &sound.Sort, &sound.Created, &sound.Updated); err != nil {
			return nil, err
		}
		out = append(out, sound)
	}

	return out, nil
}

func databaseSortSounds(db *sql.DB, categoryID int64, IDs ...int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt := `UPDATE sounds SET sort = ? WHERE category_id = ?`
	if _, err := tx.Exec(stmt, 9999, categoryID); err != nil {
		return err
	}
	for index, ID := range IDs {
		stmt = `UPDATE sounds SET sort = ? WHERE id = ? AND category_id = ?`
		if _, err := tx.Exec(stmt, index, ID, categoryID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func databaseFetchSoundByIDAndUser(db *sql.DB, ID int64, userId string) (Sound, string, error) {
	stmt := `
	SELECT s.id, s.name, s.hash, s.category_id, s.sort, s.created, s.updated, c.guild_id
	FROM sounds AS s
	INNER JOIN categories AS c
	ON s.category_id = c.id
	INNER JOIN user_guilds AS u
	ON c.guild_id = u.guild_id 
	WHERE u.user_id = ? AND s.id = ?
	`
	rows, err := db.Query(stmt, userId, ID)
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

func (s *Sound) Save(db *sql.DB) error {
	s.Updated = time.Now()
	if s.ID > 0 {
		stmt := `
		UPDATE sounds
		SET name=?, hash=?, category_id=?, sort=?, updated=?
		WHERE id=?
		`
		_, err := db.Exec(stmt, s.Name, s.Hash, s.CategoryID, s.Sort, s.Updated, s.ID)
		return err
	}

	s.Created = time.Now()
	stmt := `
	INSERT INTO sounds(name, hash, category_id, sort, created, updated) 
	VALUES(?, ?, ?, ?, ?, ?)
	`
	r, err := db.Exec(stmt, s.Name, s.Hash, s.CategoryID, s.Sort, s.Created, s.Updated)
	if err != nil {
		return err
	}
	s.ID, err = r.LastInsertId()
	return err
}

func (s *Sound) Delete(db *sql.DB) error {
	stmt := `DELETE FROM sounds WHERE id = ?`
	_, err := db.Exec(stmt, s.ID)
	// TODO: clean up unused sound files
	return err
}
