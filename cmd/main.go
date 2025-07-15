package main

import (
	"github.com/lwm-galactic/utool/pkg/app"
	"github.com/lwm-galactic/utool/pkg/cmd"
)

func main() {
	App := app.NewApp("tool", "tool",
		app.WithDescription("tool create by lwm"),
		app.WithNoConfig(),
		app.WithCommands(cmd.NewUpdateCommand(), cmd.NewPdf2DocxCommand(), cmd.NewInitCommand()),
	)
	App.Run()
}
