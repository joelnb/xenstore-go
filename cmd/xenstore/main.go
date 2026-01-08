package main

import (
	"context"
	"fmt"
	"os"
	"time"

	xenstore "github.com/joelnb/xenstore-go"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var client *xenstore.Client

func main() {
	app := &cli.Command{
		Usage:   "XenStore tools in Go",
		Version: Version,
		Metadata: map[string]interface{}{
			"compiled": time.Now(),
		},
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
			&cli.BoolFlag{
				Name:  "verbose, V",
				Usage: "More verbose output",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Output to stderr instead of stdout, could also be a file.
			log.SetOutput(os.Stderr)

			// Only log the info severity or above.
			log.SetLevel(log.InfoLevel)

			// Higher level if running in verbose mode.
			if cmd.Bool("verbose") {
				log.SetLevel(log.DebugLevel)
			}

			t, err := getTransport(cmd)
			if err != nil {
				// Returning an error here causes usage text to be printed so just exit instead
				fmt.Println(err)
				os.Exit(2)
			}

			client = xenstore.NewClient(t)

			return ctx, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			if client == nil {
				return nil
			}

			closeErr := client.Close()

			if storedErr := client.Error(); storedErr != nil {
				return storedErr
			}

			return closeErr
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:   "read",
				Flags:  []cli.Flag{},
				Usage:  "Read values from xenstore by path",
				Action: ReadCommand,
			},
			&cli.Command{
				Name:   "write",
				Flags:  []cli.Flag{},
				Usage:  "Write values to xenstore by path",
				Action: WriteCommand,
			},
			&cli.Command{
				Name:   "rm",
				Flags:  []cli.Flag{},
				Usage:  "Remove a value from xenstore by path",
				Action: RmCommand,
			},
			&cli.Command{
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
			&cli.Command{
				Name:   "vm-path",
				Flags:  []cli.Flag{},
				Usage:  "Get the path for a VM",
				Action: VMPathCommand,
			},
			&cli.Command{
				Name:   "watch",
				Flags:  []cli.Flag{},
				Usage:  "Watch a XenStore path for changes",
				Action: WatchCommand,
			},
			&cli.Command{
				Name:   "mkdir",
				Flags:  []cli.Flag{},
				Usage:  "Create path in xenstore",
				Action: MkdirCommand,
			},
			&cli.Command{
				Name:   "info",
				Flags:  []cli.Flag{},
				Usage:  "Display system information",
				Action: InfoCommand,
			},
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
