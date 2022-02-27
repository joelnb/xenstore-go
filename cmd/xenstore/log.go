package main

import (
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
)

func init() {
    setLogFormatter()

    filenameHook := filename.NewHook()
    filenameHook.Field = "source"
    filenameHook.SkipPrefixes = append(filenameHook.SkipPrefixes, "cmd/xenstore")
    log.AddHook(filenameHook)
}
