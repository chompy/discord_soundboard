package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type Table interface {
	name() string
	id() any
	init(conn *sql.DB) error
	save(conn *sql.DB) error
	delete(conn *sql.DB) error
}

var tables = []Table{
	&Guild{}, &User{}, &UserGuild{}, &Category{}, &Sound{},
}

type Client struct {
	logger *zerolog.Logger
	conn   *sql.DB
}

func New(path string, logger *zerolog.Logger) (*Client, error) {
	logger.Info().Str("databasePath", path).Msgf("Open database at %s", path)
	conn, err := sql.Open("sqlite3", path)
	return &Client{logger: logger, conn: conn}, err
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.logger.Info().Msg("Close database")
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	c.conn = nil
	return nil
}

func (c *Client) Init() error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	c.logger.Info().Msg("Initalize database")
	for _, table := range tables {
		c.logger.Info().Str("tableName", table.name()).Msgf("Create table %s", table.name())
		if err := table.init(c.conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Save(table Table) error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	c.logger.Info().Msgf("Save table %s", table.name())
	return table.save(c.conn)
}

func (c *Client) Delete(table Table) error {
	if c.conn == nil {
		return errDatabaseClosed
	}
	c.logger.Info().Any("rowID", table.id()).Str("tableName", table.name()).Msgf("Delete row %s from table %s", table.id(), table.name())
	return table.delete(c.conn)
}

func fetchById(conn *sql.DB, table Table, ID any) (*sql.Rows, error) {
	return conn.Query(fmt.Sprintf(`SELECT * FROM %s WHERE id = ? LIMIT 1`, table.name()), ID)
}

func deleteById(conn *sql.DB, table Table) error {
	_, err := conn.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table.name()), table.id())
	return err
}
