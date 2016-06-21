package push

import (
	"github.com/jeffjen/pushmoi/cmd/oauth2"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"

	"fmt"
	"os"
)

func NewListDevices() cli.Command {
	return cli.Command{
		Name:  "ls",
		Usage: "List registerd devices",
		Action: func(c *cli.Context) error {
			table := tablewriter.NewWriter(os.Stdout)

			table.SetHeader([]string{"Name", "Type", "SMS", "Active"})
			for _, dev := range oauth2.Pushbullet.Devices.Devices {
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
