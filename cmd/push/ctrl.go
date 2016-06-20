package push

import (
	"github.com/jeffjen/pushmoi/cmd/oauth2"

	"github.com/urfave/cli"

	"encoding/json"
	"fmt"
	"io"
	"os"
	path "path/filepath"
	"strings"
)

const (
	PUSH_BULLET_SETTING_FILE = "~/.pushmoi/pushbullet.setting.json"
)

var (
	PushSettings = new(Settings)
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

func NewSetCommand() cli.Command {
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
				for _, dev := range oauth2.PushBullet.Devices.Devices {
					if dev.Nickname == value {
						defer PushSettings.Dump()
						PushSettings.Default = &dev
						return nil
					}
				}
				return cli.NewExitError("Specified default target not found", 1)
			}
		},
	}
}

func NewGetCommand() cli.Command {
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
				if PushSettings.Default == nil {
					return cli.NewExitError("No default push target", 0)
				}
				fmt.Println(PushSettings.Default.Nickname)
				return nil
			}
		},
	}
}
