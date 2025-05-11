package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	xenstore "github.com/joelnb/xenstore-go"
	"github.com/urfave/cli/v3"
)

func ReadCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().First()
	if path == "" {
		return cli.Exit("Please specify the XenStore path to read", 3)
	}

	val, err := client.Read(path)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func RmCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().First()
	if path == "" {
		return cli.Exit("Please specify the XenStore path to remove", 3)
	}

	val, err := client.Remove(path)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func WriteCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().First()
	if path == "" {
		return cli.Exit("Please specify the XenStore path to write", 3)
	}

	val := cmd.Args().Get(1)
	if path == "" {
		return cli.Exit("Please specify the value to write", 3)
	}

	val, err := client.Write(path, val)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func VMPathCommand(ctx context.Context, cmd *cli.Command) error {
	domid := cmd.Args().First()
	if domid == "" {
		return cli.Exit("Please specify the domid of the VM", 3)
	}

	domidInt, err := strconv.Atoi(domid)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	path, err := client.GetDomainPath(domidInt)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(path)
	return nil
}

func ListCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().First()
	if path == "" {
		return cli.Exit("Please specify the XenStore path to list", 3)
	}

	subpaths, err := client.List(path)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	if cmd.Bool("long") {
		for _, subpath := range subpaths {
			fullpath := xenstore.JoinXenStorePath(path, subpath)

			perms, err := client.GetPermissions(fullpath)
			if err != nil {
				return cli.Exit(err.Error(), 2)
			}

			fmt.Println(fullpath, perms)
		}
	} else {
		fmt.Println(strings.Trim(fmt.Sprint(subpaths), "[]"))
	}

	return nil
}

func WatchCommand(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().First()
	if path == "" {
		return cli.Exit("Please specify the XenStore path to watch", 3)
	}

	token := cmd.Args().Get(1)
	if token == "" {
		return cli.Exit("Please specify the token to create the watch with", 3)
	}

	ch, err := client.Watch(path, token)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

OUTER:
	for {
		select {
		case rsp := <-ch:
			if err := rsp.Check(); err != nil {
				return cli.Exit(err.Error(), 2)
			} else {
				fmt.Println(rsp)
			}

		case sig := <-sigs:
			fmt.Printf("Got signal %s, removing watch and exiting!", sig)

			if err := client.UnWatch(path, token); err != nil {
				return cli.Exit(err.Error(), 2)
			}

			break OUTER
		}
	}

	return nil
}

func InfoCommand(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Socket Path:", xenstore.UnixSocketPath())
	fmt.Println("XenBus Path:", xenstore.XenBusPath())
	fmt.Println("ControlDomain:", xenstore.ControlDomain())
	fmt.Println()
	fmt.Println("Version:", Version)
	fmt.Println("GitCommit:", GitCommit)
	return nil
}
