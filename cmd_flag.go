package main

import (
	"github.com/jeffjen/pushmoi/push"

	"github.com/urfave/cli"
)

var HandlerByDriver = map[string]*cli.App{
	"pushbullet": push.MakeApp(),
}
