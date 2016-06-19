package main

import (
	"github.com/jeffjen/pushmoi/cmd"

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
		cmd.NewOAuth2Workflow(),
	}
	app.Action = func(c *cli.Context) error {
		return cli.NewExitError("Incorrect usage; Try running 'pushmoi init'", 1)
	}
	app.Run(os.Args)
}
