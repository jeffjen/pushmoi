package main

import (
	"github.com/jeffjen/pushmoi/cmd/oauth2"
	"github.com/jeffjen/pushmoi/cmd/push"

	"github.com/urfave/cli"
	"golang.org/x/net/context"

	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"
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
		push.NewListDevices(),
		push.NewSetCommand(),
		push.NewGetCommand(),
		push.NewSyncCommand(),
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "device", Value: "default", Usage: "Device to Pushbullet"},
		cli.BoolFlag{Name: "all", Usage: "Pushbullet all"},
		cli.StringFlag{Name: "type", Value: push.PUSH_NOTE_TYPE, Usage: "Push payload type"},
		cli.StringFlag{Name: "title", Usage: "Push title"},
	}
	app.Before = func(c *cli.Context) error {
		if err := oauth2.Pushbullet.Load(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		if err := push.Pushsettings.Load(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		return nil
	}
	app.Action = func(c *cli.Context) error {
		var (
			// Push configuration
			kind  = c.String("type")
			title = c.String("title")

			// Push payload configuration
			messageTmpl    = c.Args().Get(0)
			messageContext = c.Args().Get(1)

			pObj = push.NewPush(kind, title)
		)

		if messageTmpl == "" {
			return cli.NewExitError("Push message template cannot be empty", 1)
		}

		if messageContext == "-" {
			// We read the Stdin stream for the messageContext
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return cli.NewExitError("Bad message context", 1)
			}
			messageContext = string(data)
		}

		var tmpl *template.Template
		var blob interface{}
		if messageTmpl == "." {
			tmpl, _ = template.New("payload").Parse(fmt.Sprint("{{ . }}"))
			// The context is treated as is
			blob = messageContext
		} else {
			t, err := template.New("payload").ParseFiles(messageTmpl)
			if err != nil {
				return cli.NewExitError("Invalid message template", 1)
			} else {
				tmpl = t
			}
			// The context is a JSON string as argument
			err = json.Unmarshal([]byte(messageContext), blob)
			if err != nil {
				return cli.NewExitError("Invalid message context", 1)
			}
		}

		var messageOut = new(bytes.Buffer)
		if err := tmpl.Execute(messageOut, blob); err != nil {
			return cli.NewExitError("Unable to genreate message from context", 1)
		}

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		if c.Bool("all") {
			pObj.Body = messageOut.String()
			if err := pObj.Send(ctx); err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
		} else {
			dev := oauth2.Pushbullet.Has(c.String("device"))
			if dev == nil {
				dev = push.Pushsettings.Default
			}
			pObj.Iden = dev.Iden
			pObj.Body = messageOut.String()
			if err := pObj.Send(ctx); err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
		}

		return nil
	}
	app.Run(os.Args)
}
