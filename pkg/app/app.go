package app

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/lwm-galactic/utool/pkg/cli"
	"github.com/lwm-galactic/utool/pkg/log"

	"github.com/lwm-galactic/tools/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	progressMessage = color.GreenString("==>")
)

// App is model for command-line-application.
type App struct {
	// commandName: cli executable file name.
	commandName string
	// name: view for user.
	name string
	// description: executable file description.
	description string
	// options: app configuration items
	options CliOptions
	// runFunc:  cli entrance func .
	runFunc RunFunc
	// silence: -true log will not stdout recommend deploy set.
	silence bool
	// noConfig: -true --config flag will not be use you can not configuration by file.
	noConfig bool
	// commands: subcommands.
	commands []*Command

	args cobra.PositionalArgs
	cmd  *cobra.Command
}

// RunFunc defines the application's startup callback function.
type RunFunc func(option CliOptions) error

// Option defines optional parameters for initializing the application structure.
type Option func(*App)

func WithOptions(opt CliOptions) Option {
	return func(a *App) {
		a.options = opt
	}
}

// WithRunFunc is used to set the application startup function option.
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.runFunc = run
	}
}

// WithDescription is used to set the description of the application.
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

// WithSilence sets the application to silent mode, in which the program startup
// information, configuration information, and version information are not
// printed in the console.
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

// WithNoConfig set the application does not provide config flag.
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}

// WithValidArgs set the validation function to valid non-flag arguments for more information, please refer to README at PositionalArgs.
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(a *App) {
		a.args = args
	}
}

func WithCommands(cmds ...*Command) Option {
	return func(a *App) {
		a.commands = append(a.commands, cmds...)
	}
}

// WithDefaultValidArgs set default validation function to valid non-flag arguments.
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}

			return nil
		}
	}
}

// NewApp creates a new application instance based on the given application name,
// command name, and other options.
func NewApp(name string, commandName string, opts ...Option) *App {
	a := &App{
		name:        name,
		commandName: commandName,
	}

	for _, o := range opts {
		o(a)
	}

	a.buildCommand()

	return a
}
func (a *App) Build() {
	a.buildCommand()
}

func (a *App) buildCommand() {
	cmd := cobra.Command{
		Use:           FormatBaseName(a.name),
		Short:         a.name,
		Long:          a.description,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          a.args,
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true

	// init flags.
	cli.InitFlags(cmd.Flags())

	// setting subcommand
	if len(a.commands) > 0 {
		for _, command := range a.commands {
			cmd.AddCommand(command.cobraCommand())
		}
		// to add app help flag to app.
		// cmd.SetHelpCommand(helpCommand(FormatBaseName(a.commandName)))
		cmd.AddCommand(addListCmd())
	}

	if a.runFunc != nil {
		cmd.RunE = a.runCommand
	}

	var namedFlagSets cli.NamedFlagSets
	if a.options != nil {
		namedFlagSets = a.options.Flags()
	}

	// add config flag
	if !a.noConfig {
		addConfigFlag(a.commandName, namedFlagSets.FlagSet("global"))
	}
	// cli.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	cmd.Flags().AddFlagSet(namedFlagSets.FlagSet("global"))

	addCmdTemplate(&cmd, namedFlagSets)
	a.cmd = &cmd
}

// Run is used to launch the application.
func (a *App) Run() {
	if err := a.cmd.Execute(); err != nil {
		fmt.Printf("%v %v\n", color.RedString("Error:"), err)
		os.Exit(1)
	}
}

// to run app.
func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	printWorkingDir()
	cli.InitFlags(cmd.Flags())
	if !a.noConfig {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}
	if !a.silence {
		log.Infof("%v Starting %s ...", progressMessage, a.name)

		if !a.noConfig {
			log.Infof("%v Config file used: `%s`", progressMessage, viper.ConfigFileUsed())
		}
	}
	if a.options != nil {
		if err := a.applyOptionRules(); err != nil {
			return err
		}
	}
	// run application
	if a.runFunc != nil {
		return a.runFunc(a.options)
	}
	return nil
}

func (a *App) applyOptionRules() error {
	if completableOptions, ok := a.options.(CompletableOptions); ok {
		if err := completableOptions.Complete(); err != nil {
			return err
		}
	}

	if errs := a.options.Validate(); len(errs) != 0 {
		return fmt.Errorf("invalid options: %v", errs)
	}

	if printableOptions, ok := a.options.(PrintableOptions); ok && !a.silence {
		log.Infof("%v Config: `%s`", progressMessage, printableOptions.String())
	}

	return nil
}

// terminal beautify
func addCmdTemplate(cmd *cobra.Command, namedFlagSets cli.NamedFlagSets) {
	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cli.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)

		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cli.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

}

func addListCmd() *cobra.Command {
	// 创建 list 子命令
	return &cobra.Command{
		Use:   "list",
		Short: "List all available commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available commands:")
			for _, c := range cmd.Root().Commands() {
				fmt.Printf(" - %s: %s\n", c.Name(), c.Short)
			}
			fmt.Println("To Use tool subcommands --help to show how subcommand use")
			return nil
		},
	}
}

// to show working dir
func printWorkingDir() {
	wd, _ := os.Getwd()
	log.Infof("%v WorkingDir: %s", progressMessage, wd)
}
