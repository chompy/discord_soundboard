package main

import (
	"database/sql"
	"log"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

const databaseFilename = "database.sqlite"

func databaseOpen(config *Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path.Join(config.StoragePath, databaseFilename))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func databaseInit(config *Config) error {
	log.Println("> Initalize database.")
	db, err := databaseOpen(config)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := databaseCreateUserTable(db); err != nil {
		return err
	}
	if err := databaseCreateCategoryTable(db); err != nil {
		return err
	}
	if err := databaseCreateSoundTable(db); err != nil {
		return err
	}
	if err := databaseCreateGuildTable(db); err != nil {
		return err
	}
	if err := databaseCreateUserGuildTable(db); err != nil {
		return err
	}
	log.Println("  - Done")
	return nil
}
