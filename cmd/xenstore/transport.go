//go:build !windows
// +build !windows

package main

import (
	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli/v3"
)

func getTransport(cmd *cli.Command) (xenstore.Transport, error) {
	if cmd.Bool("use-socket") {
		var sockPath string
		if cmd.IsSet("socket-path") {
			sockPath = cmd.String("socket-path")
		} else {
			sockPath = xenstore.UnixSocketPath()
		}

		return xenstore.NewUnixSocketTransport(sockPath)
	} else {
		var devPath string
		if cmd.IsSet("xenbus-path") {
			devPath = cmd.String("xenbus-path")
		} else {
			devPath = xenstore.XenBusPath()
		}

		return xenstore.NewXenBusTransport(devPath)
	}
}
