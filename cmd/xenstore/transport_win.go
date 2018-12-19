// +build windows

package main

import (
	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli"
)

func getTransport(ctx *cli.Context) (xenstore.Transport, error) {
	return nil, xenstore.NewWinPVTransport()
}
