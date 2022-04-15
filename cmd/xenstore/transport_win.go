//go:build windows
// +build windows

package main

import (
	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli/v2"
)

func getTransport(ctx *cli.Context) (xenstore.Transport, error) {
	return xenstore.NewWinPVTransport()
}
