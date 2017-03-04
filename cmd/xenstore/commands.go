package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joelnb/xenstore-go"
	"gopkg.in/urfave/cli.v2"
)

func ReadCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.Exit("please specify the XenStore path to read", 3)
	}

	val, err := client.Read(path)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func RmCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.Exit("please specify the XenStore path to remove", 3)
	}

	val, err := client.Remove(path)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func WriteCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.Exit("please specify the XenStore path to write", 3)
	}

	val := ctx.Args().Get(1)
	if path == "" {
		return cli.Exit("please specify the value to write", 3)
	}

	val, err := client.Write(path, val)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	fmt.Println(val)
	return nil
}

func ListCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.Exit("please specify the XenStore path to read", 3)
	}

	subpaths, err := client.List(path)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	if ctx.Bool("long") {
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

func WatchCommand(ctx *cli.Context) error {
	path := ctx.Args().First()
	if path == "" {
		return cli.Exit("please specify the XenStore path to watch", 3)
	}

	token := ctx.Args().Get(1)
	if token == "" {
		return cli.Exit("please specify the token to create the watch with", 3)
	}

	ch, err := client.Watch(path, token)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	for {
		rsp := <-ch

		if err := rsp.Check(); err != nil {
			return cli.Exit(err.Error(), 2)
		} else {
			// decoded, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(string(rsp.Payload))
			// if err != nil {
			//  continue
			//  //return cli.Exit(fmt.Sprintf("decode error: %s", err), 7)
			// }
			// rsp.Payload = decoded

			rspJson, err := json.Marshal(rsp)
			if err != nil {
				return cli.Exit(err.Error(), 2)
			}

			fmt.Println(string(rspJson))
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
