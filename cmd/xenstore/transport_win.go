//go:build windows
// +build windows

package main

import (
	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli/v3"
)

func getTransport(cmd *cli.Command) (xenstore.Transport, error) {
	return xenstore.NewWinPVTransport()
}
