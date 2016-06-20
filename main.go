package main

import (
	"github.com/jeffjen/pushmoi/cmd/oauth2"

	"github.com/urfave/cli"

	"os"
)

func main() {
	app := cli.NewApp()

	app.Version = "0.0.1"
	app.Name = "pushmoi"
	app.Usage = "Send/Receive Pushbullet message"
	app.Authors = []cli.Author{
		cli.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
	app.Commands = []cli.Command{
		oauth2.NewOAuth2Workflow(),
	}
	app.Before = func(c *cli.Context) error {
		if err := oauth2.PushBullet.Load(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		} else {
			return nil
		}
	}
	app.Action = func(c *cli.Context) error {
		return cli.NewExitError("Incorrect usage; Try running 'pushmoi init'", 1)
	}
	app.Run(os.Args)
}
