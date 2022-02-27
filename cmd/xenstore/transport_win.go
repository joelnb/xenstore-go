//go:build windows
// +build windows

package main

import (
	"errors"

	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli"
)

func getTransport(ctx *cli.Context) (xenstore.Transport, error) {
	err := xenstore.NewWinPVTransport()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("Not yet implemented")
}
