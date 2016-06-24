package push

import (
	"github.com/jeffjen/pushmoi/oauth2"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"golang.org/x/net/context"

	"encoding/json"
	"fmt"
	"io"
	"os"
	path "path/filepath"
	"strings"
	"time"
)

const (
	PUSH_BULLET_SETTING_FILE = "~/.pushmoi/pushbullet.setting.json"
)

var (
	Pushsettings = new(Settings)
)

type Settings struct {
	Default *oauth2.Device `json:"default"`
}

func (s *Settings) Load() error {
	conf, err := getConfigPath()
	if err != nil {
		return err
	}
	origin, err := os.OpenFile(conf, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer origin.Close()
	if err = json.NewDecoder(origin).Decode(s); err == io.EOF {
		return nil
	} else {
		return err
	}
}

func (s *Settings) Dump() error {
	conf, err := getConfigPath()
	if err != nil {
		return err
	}
	origin, err := os.OpenFile(conf, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer origin.Close()
	return json.NewEncoder(origin).Encode(s)
}

func getConfigPath() (string, error) {
	conf := strings.Replace(PUSH_BULLET_SETTING_FILE, "~", os.Getenv("HOME"), 1)
	confdir := path.Dir(conf)
	if _, err := os.Stat(confdir); err != nil {
		if os.IsNotExist(err) {
			return conf, os.MkdirAll(confdir, 0700)
		} else {
			return "", err
		}
	} else {
		return conf, nil
	}
}

func BeforeAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if err := oauth2.Pushbullet.Load(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		if err := Pushsettings.Load(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		return nil
	}
}

func NewCommand() cli.Command {
	return cli.Command{
		Name:   "pushbullet",
		Usage:  "Pushbullet configuration",
		Before: BeforeAction(),
		Subcommands: []cli.Command{
			newListDevices(),
			newSetCommand(),
			newGetCommand(),
			newSyncCommand(),
		},
	}
}

func newSetCommand() cli.Command {
	return cli.Command{
		Name:  "set",
		Usage: "Configure settings in pushmoi",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				return cli.NewExitError("Invalid use; must have [setting] [value]", 1)
			}
			setting, value := c.Args().Get(0), c.Args().Get(1)
			if value == "" {
				return cli.NewExitError("Invalid use; must have [setting] [value]", 1)
			}
			switch setting {
			default:
				return cli.NewExitError("Invalid use; setting not found", 1)
			case "default":
				for _, dev := range oauth2.Pushbullet.Devices {
					if dev.Nickname == value {
						defer Pushsettings.Dump()
						Pushsettings.Default = dev
						return nil
					}
				}
				return cli.NewExitError("Specified default target not found", 1)
			}
		},
	}
}

func newGetCommand() cli.Command {
	return cli.Command{
		Name:  "get",
		Usage: "Retrieve settings in pushmoi",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				return cli.NewExitError("Invalid use; must have [setting]", 1)
			}
			setting := c.Args().Get(0)
			switch setting {
			default:
				return cli.NewExitError("Invalid use; setting not found", 1)
			case "default":
				if Pushsettings.Default == nil {
					return cli.NewExitError("No default push target", 0)
				}
				fmt.Println(Pushsettings.Default.Nickname)
				return nil
			}
		},
	}
}

func newSyncCommand() cli.Command {
	return cli.Command{
		Name:  "sync",
		Usage: "Sync config and check settings validity",
		Action: func(c *cli.Context) error {
			defer oauth2.Pushbullet.Dump()

			ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)

			// Sync user profile and registered devices
			if err := oauth2.Pushbullet.Sync(ctx); err != nil {
				return cli.NewExitError("Failed to sync Pushbullet info", 3)
			} else {
				return nil
			}
		},
	}
}

func newListDevices() cli.Command {
	return cli.Command{
		Name:  "ls",
		Usage: "List registerd devices",
		Action: func(c *cli.Context) error {
			table := tablewriter.NewWriter(os.Stdout)

			table.SetHeader([]string{"Name", "Type", "SMS", "Active"})
			for _, dev := range oauth2.Pushbullet.Devices {
				if dev.Nickname == "" {
					continue
				}
				table.Append([]string{dev.Nickname, dev.Icon, fmt.Sprint(dev.HasSms), fmt.Sprint(dev.Active)})
			}
			table.Render()

			return nil
		},
	}
}
