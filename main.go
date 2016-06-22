package main

import (
	"github.com/jeffjen/pushmoi/oauth2"
	"github.com/jeffjen/pushmoi/push"

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
		cli.StringFlag{Name: "email", Value: "", Usage: "Pushbullet to email owner"},
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
			all    = c.Bool("all")
			email  = c.String("email")
			device = c.String("device")

			// Push configuration
			kind  = c.String("type")
			title = c.String("title")

			// Push payload configuration
			messageTmpl    = c.Args().Get(0)
			messageContext = c.Args().Get(1)

			// Blog for holding message context
			blob interface{} = make(map[string]interface{})

			// Template for outbound message
			tmpl *template.Template

			// Pushbullet
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

		if messageTmpl == "." {
			tmpl, _ = template.New("payload").Parse(fmt.Sprint("{{ . }}"))
			// The context is treated as is
			blob = messageContext
		} else {
			buf, err := ioutil.ReadFile(messageTmpl)
			if err != nil {
				return cli.NewExitError("Invalid message template", 1)
			}
			t, err := template.New("payload").Parse(string(buf))
			if err != nil {
				return cli.NewExitError("Invalid message template", 1)
			} else {
				tmpl = t
			}
			// The context is a JSON string as argument
			err = json.Unmarshal([]byte(messageContext), &blob)
			if err != nil {
				// The context is treated as is
				blob = messageContext
			}
		}

		var messageOut = new(bytes.Buffer)
		if err := tmpl.Execute(messageOut, blob); err != nil {
			fmt.Println(err)
			return cli.NewExitError("Unable to genreate message from context", 1)
		}
		pObj.Body = messageOut.String()

		switch {
		default:
			dev := oauth2.Pushbullet.Has(device)
			if dev == nil {
				dev = push.Pushsettings.Default
			}
			pObj.Iden = dev.Iden
			break
		case all == true:
			break
		case email != "":
			pObj.Email = email
			break
		}

		for attempts := 0; attempts < 3; attempts++ {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			if err := pObj.Send(ctx); err != nil {
				return cli.NewExitError(err.Error(), 1)
			} else {
				break // success!
			}
		}

		return nil
	}
	app.Run(os.Args)
}
