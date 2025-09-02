package main

import (
	"database/sql"
	"log"
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

func databaseCreateCategoryTable(db *sql.DB) error {
	log.Println("  - Create categories table")
	stmt := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		guild_id TEXT,
		sort INTEGER,
		created DATETIME,
		updated DATETIME
	)
	`
	_, err := db.Exec(stmt)
	return err
}

func databaseFetchCategoryByID(db *sql.DB, ID int64) (Category, error) {
	category := Category{}
	stmt := "SELECT * FROM categories WHERE id = ?"
	rows, err := db.Query(stmt, ID)
	if err != nil {
		return category, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&category.ID, &category.Name, &category.GuildID, &category.Sort, &category.Created, &category.Updated)
		return category, err
	}
	return category, nil
}

func databaseFetchCategoriesByGuildID(db *sql.DB, guildID string) ([]Category, error) {
	out := make([]Category, 0)

	stmt := "SELECT * FROM categories WHERE guild_id = ? ORDER BY sort, id ASC"
	rows, err := db.Query(stmt, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		category := Category{}
		if err := rows.Scan(&category.ID, &category.Name, &category.GuildID, &category.Sort, &category.Created, &category.Updated); err != nil {
			return nil, err
		}
		out = append(out, category)
	}

	return out, nil
}

func databaseDeleteCategoryByID(db *sql.DB, ID int64) error {
	stmt := `DELETE FROM sounds WHERE category_id = ?`
	_, err := db.Exec(stmt, ID)
	if err != nil {
		return err
	}

	stmt = `DELETE FROM categories WHERE id = ?`
	_, err = db.Exec(stmt, ID)
	return err
}

func databaseSortCategories(db *sql.DB, guildID string, IDs ...int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt := `UPDATE categories SET sort = ? WHERE guild_id = ?`
	if _, err := tx.Exec(stmt, 9999, guildID); err != nil {
		return err
	}
	for index, ID := range IDs {
		stmt = `UPDATE categories SET sort = ? WHERE id = ? AND guild_id = ?`
		if _, err := tx.Exec(stmt, index, ID, guildID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Category) Save(db *sql.DB) error {
	c.Updated = time.Now()
	if c.ID > 0 {
		stmt := `
		UPDATE categories
		SET name=?, guild_id=?, sort=?, updated=?
		WHERE id=?
		`
		_, err := db.Exec(stmt, c.Name, c.GuildID, c.Sort, c.Updated, c.ID)
		return err
	}

	c.Created = time.Now()
	stmt := `
	INSERT INTO categories(name, guild_id, sort, created, updated) 
	VALUES(?, ?, ?, ?, ?)
	`
	r, err := db.Exec(stmt, c.Name, c.GuildID, c.Sort, c.Created, c.Updated)
	if err != nil {
		return err
	}
	c.ID, err = r.LastInsertId()
	return err
}

func (c *Category) Delete(db *sql.DB) error {
	return databaseDeleteCategoryByID(db, c.ID)
}
