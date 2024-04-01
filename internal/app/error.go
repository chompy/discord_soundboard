package app

import "errors"

var errNoToken = errors.New("missing discord bot token")
var errMissingParam = errors.New("missing one or more required parameters")
var errSoundNotFound = errors.New("sound not found")
