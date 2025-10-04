package web

import "errors"

var (
	errNotAuthenticated = errors.New("not authenticated")
	errNotAuthorized    = errors.New("not authorized")
	errInvalidAuth      = errors.New("missing or invalid data during auth")
	errUnknown          = errors.New("an unspecified error occurred")
	errMissingParam     = errors.New("missing one or more required parameters")
	errInvalidParam     = errors.New("invalid parameter(s)")
	errNotFound         = errors.New("requested resource was not found")
	errNoActiveChannel  = errors.New("user is not in a voice channel")
	errInvalidMethod    = errors.New("invalid method")
	errInvalidData      = errors.New("invalid data")
	errDatabaseWrite    = errors.New("unable to write database")
	errDatabaseRead     = errors.New("unable to read database")
	errDiscordApi       = errors.New("failed to access discord api")
	errSoundRead        = errors.New("unable to read sound")
	errSoundWrite       = errors.New("unable to save sound")
)

// err to http status code map
var errHttpStatusCodeMap = map[error]int{
	errNotAuthenticated: 401,
	errNotAuthorized:    401,
	errMissingParam:     400,
	errInvalidMethod:    400,
	errNotFound:         404,
	errNoActiveChannel:  400,
}
