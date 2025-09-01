package main

import "errors"

// application configuration errors
var errNoToken = errors.New("missing discord bot token")

// auth errors
var errNotAuthenticated = errors.New("not authenticated")
var errNotAuthorized = errors.New("not authorized")

// user errors
var errMissingParam = errors.New("missing one or more required parameters")
var errNotFound = errors.New("requested resource was not found")
var errNoAvailableGuilds = errors.New("no available servers")

var errInvalidMethod = errors.New("invalid method")
var errInvalidData = errors.New("invalid data")
var errInvalidSound = errors.New("invalid sound data")
var errInvalidInstruction = errors.New("invalid multi sound instruction")
var errInvalidAuth = errors.New("missing or invalid data during auth")

var errDatabaseWrite = errors.New("unable to write database")
var errDatabaseRead = errors.New("unable to read database")
