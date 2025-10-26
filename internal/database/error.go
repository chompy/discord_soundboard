package database

import "errors"

var (
	errDatabaseClosed = errors.New("database is closed")
	errRecordNotFound = errors.New("record not found")
)

func IsRecordNotFoundError(err error) bool {
	return err == errRecordNotFound
}
