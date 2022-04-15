//go:build !windows
// +build !windows

package main

import (
	log "github.com/sirupsen/logrus"
)

func setLogFormatter() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}
