package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joelnb/xenstore-go"
	"github.com/urfave/cli"
)

var client *xenstore.Client

func main() {
	app := &cli.App{
		Usage:    "XenStore tools in Go",
		Version:  Version,
		Compiled: time.Now(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "socket-path",
				Usage: "Path to the xenstore unix socket",
			},
			&cli.StringFlag{
				Name:  "xenbus-path",
				Usage: "Path to the xenbus device",
			},
			&cli.BoolFlag{
				Name:  "use-socket, s",
				Usage: "Use the socket rather than the xenbus device",
			},
		},
		Before: func(ctx *cli.Context) error {
			t, err := getTransport(ctx)
			if err != nil {
				// Returning an error here causes usage text to be printed so just exit instead
				fmt.Println(err)
				os.Exit(2)
			}

			client = xenstore.NewClient(t)

			return nil
		},
		After: func(ctx *cli.Context) error {
			client.Close()
			return nil
		},
		Commands: []cli.Command{
			cli.Command{
				Name:   "read",
				Flags:  []cli.Flag{},
				Usage:  "Read values from xenstore by path",
				Action: ReadCommand,
			},
			cli.Command{
				Name:   "write",
				Flags:  []cli.Flag{},
				Usage:  "Write values to xenstore by path",
				Action: WriteCommand,
			},
			cli.Command{
				Name:   "rm",
				Flags:  []cli.Flag{},
				Usage:  "Remove a value from xenstore by path",
				Action: RmCommand,
			},
			cli.Command{
				Name:    "list",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "long, l",
					},
				},
				Usage:  "List values from xenstore by path",
				Action: ListCommand,
			},
			cli.Command{
				Name:   "watch",
				Flags:  []cli.Flag{},
				Usage:  "Watch a XenStore path for changes",
				Action: WatchCommand,
			},
			cli.Command{
				Name:   "info",
				Flags:  []cli.Flag{},
				Usage:  "Display system information",
				Action: InfoCommand,
			},
		},
	}

	app.Run(os.Args)
}
