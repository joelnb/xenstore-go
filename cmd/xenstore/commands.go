package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	xenstore "github.com/joelnb/xenstore-go"
	"github.com/urfave/cli"
)

func ReadCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.NewExitError("Please specify the XenStore path to read", 3)
	}

	val, err := client.Read(path)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func RmCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.NewExitError("Please specify the XenStore path to remove", 3)
	}

	val, err := client.Remove(path)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func WriteCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.NewExitError("Please specify the XenStore path to write", 3)
	}

	val := ctx.Args().Get(1)
	if path == "" {
		return cli.NewExitError("Please specify the value to write", 3)
	}

	val, err := client.Write(path, val)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func VMPathCommand(ctx *cli.Context) error {
	domid := ctx.Args().First()
	if domid == "" {
		return cli.NewExitError("Please specify the domid of the VM", 3)
	}

	domidInt, err := strconv.Atoi(domid)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	path, err := client.GetDomainPath(domidInt)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	fmt.Println(path)
	return nil
}

func ListCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.NewExitError("Please specify the XenStore path to list", 3)
	}

	subpaths, err := client.List(path)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	if ctx.Bool("long") {
		for _, subpath := range subpaths {
			fullpath := xenstore.JoinXenStorePath(path, subpath)

			perms, err := client.GetPermissions(fullpath)
			if err != nil {
				return cli.NewExitError(err.Error(), 2)
			}

			fmt.Println(fullpath, perms)
		}
	} else {
		fmt.Println(strings.Trim(fmt.Sprint(subpaths), "[]"))
	}

	return nil
}

func WatchCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.NewExitError("Please specify the XenStore path to watch", 3)
	}

	token := ctx.Args().Get(1)
	if token == "" {
		return cli.NewExitError("Please specify the token to create the watch with", 3)
	}

	ch, err := client.Watch(path, token)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

OUTER:
	for {
		select {
		case rsp := <-ch:
			if err := rsp.Check(); err != nil {
				return cli.NewExitError(err.Error(), 2)
			} else {
				fmt.Println(rsp)
			}

		case sig := <-sigs:
			fmt.Printf("Got signal %s, removing watch and exiting!", sig)

			if err := client.UnWatch(path, token); err != nil {
				return cli.NewExitError(err.Error(), 2)
			}

			break OUTER
		}
	}

	return nil
}

func InfoCommand(ctx *cli.Context) error {
	fmt.Println("Socket Path:", xenstore.UnixSocketPath())
	fmt.Println("XenBus Path:", xenstore.XenBusPath())
	fmt.Println("ControlDomain:", xenstore.ControlDomain())
	fmt.Println("Version:", Version)
	return nil
}
