package discord

import "errors"

var (
	errClientClosed    = errors.New("discord client has been closed")
	errNoActiveChannel = errors.New("user is not in a voice channel")
)
