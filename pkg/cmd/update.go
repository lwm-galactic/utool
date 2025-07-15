package cmd

import (
	"fmt"
	"github.com/lwm-galactic/utool/pkg/app"
)

func NewUpdateCommand() *app.Command {
	return app.NewCommand("update", "to update tool", app.WithCommandRunFunc(run))
}

func run(option app.CliOptions) error {
	fmt.Println("update called")

	return nil
}
