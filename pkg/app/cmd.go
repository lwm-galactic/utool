package app

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/lwm-galactic/utool/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
)

// Command is a sub command structure of a cli application.
// It is  recommended that a command be created with the app.NewCommand().
type Command struct {
	usage    string
	desc     string
	options  CliOptions
	commands []*Command
	runFunc  RunFunc
}

// NewCommand creates a new sub command instance based on the given command name and other options.
func NewCommand(usage string, desc string, opts ...CommandOption) *Command {
	c := &Command{
		usage: usage,
		desc:  desc,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

// CommandOption defines optional parameters for initializing the sub command structure.
type CommandOption func(*Command)

// WithCommandRunFunc functional options pattern to set RunCommandFunc
func WithCommandRunFunc(run RunFunc) CommandOption {
	return func(c *Command) {
		c.runFunc = run
	}
}

func WithCommandOptions(opt CliOptions) CommandOption {
	return func(c *Command) {
		c.options = opt
	}
}

// FormatBaseName is formatted as an executable file name under different
// operating systems according to the given name.
func FormatBaseName(basename string) string {
	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}

	return basename
}

// conversion customize Command  to cobra.Command
func (c *Command) cobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   c.usage,
		Short: c.desc,
	}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.Flags().SortFlags = false
	if len(c.commands) > 0 {
		for _, command := range c.commands {
			cmd.AddCommand(command.cobraCommand())
		}
	}
	if c.runFunc != nil {
		cmd.RunE = c.runCommand
	}
	if c.options != nil {
		for _, f := range c.options.Flags().FlagSets {
			cmd.Flags().AddFlagSet(f)
		}
	}

	// to add --help flag to command
	addHelpCommandFlag(c.usage, cmd.Flags())

	return cmd
}

func (c *Command) runCommand(cmd *cobra.Command, args []string) error {
	cli.InitFlags(cmd.Flags())
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}
	err = viper.Unmarshal(c.options)
	if err != nil {
		return err
	}
	if c.runFunc != nil {
		if err := c.runFunc(c.options); err != nil {
			fmt.Printf("%v %v\n", color.RedString("Error:"), err)
			os.Exit(1)
		}
	}
	return nil
}
