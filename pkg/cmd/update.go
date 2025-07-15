package cmd

import (
	"github.com/lwm-galactic/utool/pkg/app"
	"github.com/lwm-galactic/utool/pkg/log"
)

func NewUpdateCommand() *app.Command {
	return app.NewCommand("update", "to update tool", app.WithCommandRunFunc(run))
}

func run(option app.CliOptions) error {
	log.Info("update called")

	log.Info("update called success")
	return nil
}
