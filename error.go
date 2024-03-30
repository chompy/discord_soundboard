package main

import "errors"

var errNoToken = errors.New("missing discord bot token")
var errMissingParam = errors.New("missing one or more required parameters")
