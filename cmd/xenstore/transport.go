// +build !windows

package main

import (
	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli"
)

func getTransport(ctx *cli.Context) (xenstore.Transport, error) {
	if ctx.Bool("use-socket") {
		var sockPath string
		if ctx.IsSet("socket-path") {
			sockPath = ctx.String("socket-path")
		} else {
			sockPath = xenstore.UnixSocketPath()
		}

		return xenstore.NewUnixSocketTransport(sockPath)
	} else {
		var devPath string
		if ctx.IsSet("xenbus-path") {
			devPath = ctx.String("xenbus-path")
		} else {
			devPath = xenstore.XenBusPath()
		}

		return xenstore.NewXenBusTransport(devPath)
	}
}
