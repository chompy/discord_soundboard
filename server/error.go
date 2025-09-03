package main

import "errors"

// application configuration errors
var errMissingBotToken = errors.New("missing discord bot token")
var errMissingAuthConfig = errors.New("missing discord auth client id and/or client secret")

// auth errors
var errNotAuthenticated = errors.New("not authenticated")
var errNotAuthorized = errors.New("not authorized")
var errInvalidAuth = errors.New("missing or invalid data during auth")

// discord errors
var errDiscordApi = errors.New("failed to access discord api")

// database errors
var errDatabaseOpen = errors.New("unable to access database")
var errDatabaseWrite = errors.New("unable to write database")
var errDatabaseRead = errors.New("unable to read database")

// sound storage errors
var errSoundRead = errors.New("unable to read sound")
var errSoundWrite = errors.New("unable to save sound")

// user errors
var errMissingParam = errors.New("missing one or more required parameters")
var errInvalidParam = errors.New("invalid parameter(s)")
var errNotFound = errors.New("requested resource was not found")
var errNoActiveChannel = errors.New("user is not in a voice channel")
var errInvalidMethod = errors.New("invalid method")
var errInvalidData = errors.New("invalid data")

var errUnknown = errors.New("an unspecified error occurred")

// err to http status code map
var errHttpStatusCodeMap = map[error]int{
	errNotAuthenticated: 401,
	errNotAuthorized:    401,
	errMissingParam:     400,
	errInvalidMethod:    400,
	errNotFound:         404,
	errNoActiveChannel:  400,
}
