package main

import (
	"github.com/jeffjen/pushmoi/oauth2"
	"github.com/jeffjen/pushmoi/push"

	"github.com/urfave/cli"

	"fmt"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Version = "0.1.0"
	app.Name = "pushmoi"
	app.Usage = "Send or Receive message"
	app.Authors = []cli.Author{
		cli.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
	app.Commands = []cli.Command{
		oauth2.NewOAuth2Workflow(),
		push.NewCommand(),

		// Command for actually push a message
		sendCommand(),
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "driver", Value: "pushbullet", EnvVar: "PUSHMOI_DRIVER", Usage: "Driver to use for push message"},
	}
	app.Before = func(c *cli.Context) error {
		name := c.Args().First()
		cmd := c.App.Command(name)
		if cmd == nil {
			return nil
		}
		switch cmd.Name {
		default:
			break
		case "send":
			cli.HandleExitCoder(processSend(c))
			os.Exit(0)
		}
		return nil
	}
	app.Action = func(c *cli.Context) error {
		var errMsg = fmt.Sprintf("Invalid usage; Try running pushmoi init")
		return cli.NewExitError(errMsg, 1)
	}
	app.Run(os.Args)
}

func processSend(c *cli.Context) error {
	handler, ok := HandlerByDriver[c.String("driver")]
	if !ok {
		return cli.NewExitError("", 1)
	}
	handler.Name = "pushmoi send"
	handler.Usage = "Send a message through driver"
	handler.UsageText = `pushmoi [global options] send [. | template] [- | message]`
	handler.HideHelp = true
	handler.HideVersion = true
	return handler.Run(c.Args())
}

func sendCommand() cli.Command {
	return cli.Command{
		Name:      "send",
		Usage:     "Send a message through driver",
		UsageText: `pushmoi [global options] [. | template] [- | message]`,
	}
}
